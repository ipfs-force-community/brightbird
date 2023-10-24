package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/venus/venus-shared/actors/policy"
	clientapi "github.com/filecoin-project/venus/venus-shared/api/market/client"
	vtypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/filecoin-project/venus/venus-shared/types/market/client"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-cidutil/cidenc"
	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multibase"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletclient "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-client"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("init-offline-order")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "init-offline-order",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "droplet-client发起离线订单",
}

type TestCaseParams struct {
	DropletClient dropletclient.DropletClientDeployReturn `json:"DropClient" jsonschema:"DropClient" title:"DropletClient" description:"droplet client return"`

	File    string `json:"File" jsonschema:"File" title:"File"`
	CarFile string `json:"CarFile" jsonschema:"CarFile" title:"CarFile"`

	MinerAddress  address.Address `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true"`
	Price         vtypes.FIL      `json:"Price"  jsonschema:"Price"  title:"Price" require:"true" default:"0.01fil" description:"price"`
	Duration      int64           `json:"Duration"  jsonschema:"Duration"  title:"Duration" default:"518400" require:"true" description:"Set the price of the ask for retrievals"`
	StatelessDeal bool            `json:"StatelessDeal"  jsonschema:"StatelessDeal"  title:"StatelessDeal" default:"false" require:"true" description:"true离线订单/false在线订单"`
	From          address.Address `json:"From"  jsonschema:"From"  title:"From" require:"false" description:"From"`
	StartEpoch    int64           `json:"StartEpoch"  jsonschema:"StartEpoch"  title:"StartEpoch" default:"-1" require:"false" description:"StartEpoch"`
}

type InitOfflineOrderReturn struct {
	ProposalCid string
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*InitOfflineOrderReturn, error) {
	api, closer, err := clientapi.DialIMarketClientRPC(ctx, params.DropletClient.SvcEndpoint.ToMultiAddr(), params.DropletClient.ClientToken, nil)
	if err != nil {
		return nil, fmt.Errorf("dial market client rpc failed")
	}
	defer closer()

	dataImportReturns, err := GetCidAndPieceSize(ctx, api, params.File, params.CarFile)
	if err != nil {
		return nil, err
	}

	proposalCid, err := StorageDealsInit(ctx, params, api, dataImportReturns)
	if err != nil {
		return nil, err
	}

	return &InitOfflineOrderReturn{
		ProposalCid: proposalCid,
	}, nil
}

type DataImportReturn struct {
	Root      cid.Cid
	Cid       cid.Cid
	PieceSize abi.UnpaddedPieceSize
}

func GetCidAndPieceSize(ctx context.Context, api clientapi.IMarketClient, file, carFile string) (*DataImportReturn, error) {
	absPath, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}

	ref := client.FileRef{
		Path:  absPath,
		IsCAR: false,
	}
	c, err := api.ClientImport(ctx, ref)
	if err != nil {
		return nil, err
	}

	ref2 := client.FileRef{
		Path:  file,
		IsCAR: false,
	}

	err = api.ClientGenCar(ctx, ref2, carFile)
	if err != nil {
		return nil, err
	}

	ret, err := api.ClientCalcCommP(ctx, carFile)
	if err != nil {
		return nil, err
	}

	return &DataImportReturn{
		Root:      c.Root,
		Cid:       ret.Root,
		PieceSize: ret.Size,
	}, nil
}

func StorageDealsInit(ctx context.Context, params TestCaseParams, api clientapi.IMarketClient, initData *DataImportReturn) (string, error) {
	data := &storagemarket.DataRef{
		TransferType: storagemarket.TTManual,
		Root:         initData.Root,
		PieceCid:     &initData.Cid,
		PieceSize:    initData.PieceSize,
	}

	MinDealDuration, MaxDealDuration := policy.DealDurationBounds(0)
	if abi.ChainEpoch(params.Duration) < MinDealDuration {
		return "", fmt.Errorf("minimum deal duration is %d blocks", MinDealDuration)
	}
	if abi.ChainEpoch(params.Duration) > MaxDealDuration {
		return "", fmt.Errorf("maximum deal duration is %d blocks", MaxDealDuration)
	}

	var wallet address.Address
	if params.From != address.Undef {
		wallet = params.From
	} else {
		def, err := api.DefaultAddress(ctx)
		if err != nil {
			return "", err
		}
		wallet = def
	}

	log.Infoln("data:", data)
	log.Infoln("wallet:", wallet)
	log.Infoln("params.MinerAddress:", params.MinerAddress)
	log.Infoln("EpochPrice:", vtypes.BigInt(params.Price))
	log.Infoln("MinBlocksDuration:", uint64(params.Duration))
	log.Infoln("DealStartEpoch:", abi.ChainEpoch(params.StartEpoch))

	sdParams := &client.DealParams{
		Data:               data,
		Wallet:             wallet,
		Miner:              params.MinerAddress,
		EpochPrice:         vtypes.BigInt(params.Price),
		MinBlocksDuration:  uint64(params.Duration),
		DealStartEpoch:     abi.ChainEpoch(params.StartEpoch),
		FastRetrieval:      true,
		VerifiedDeal:       false,
		ProviderCollateral: big.NewInt(0),
	}

	var proposal *cid.Cid
	var err error
	if params.StatelessDeal {
		if params.Price.Int64() != 0 {
			return "", fmt.Errorf("when manual-stateless-deal is enabled, you must also provide a 'price' of 0 and specify 'manual-piece-cid' and 'manual-piece-size'")
		}
		proposal, err = api.ClientStatelessDeal(ctx, sdParams)
	} else {
		proposal, err = api.ClientStartDeal(ctx, sdParams)
	}
	if err != nil {
		return "", err
	}

	encoder := cidenc.Encoder{Base: multibase.MustNewEncoder(multibase.Base32)}
	if err != nil {
		return "", err
	}

	log.Infoln("DealCid cid: ", encoder.Encode(*proposal))
	return encoder.Encode(*proposal), nil
}
