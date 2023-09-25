package main

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"text/template"

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
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/pvc"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
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
	Auth          sophonauth.SophonAuthDeployReturn       `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Venus         venus.VenusDeployReturn                 `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	DropletClient dropletclient.DropletClientDeployReturn `json:"DropClient" jsonschema:"DropClient" title:"DropletClient" description:"droplet client return"`
	PieceStore    pvc.PvcReturn                           `json:"PieceStore" jsonschema:"PieceStore" title:"PieceStore" require:"true" description:"piece storage"`

	MinerAddress address.Address `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true"`
	Price        vtypes.FIL      `json:"Price"  jsonschema:"Price"  title:"Price" require:"true" default:"0.01fil" description:"price"`
	Duration     int64           `json:"Duration"  jsonschema:"Duration"  title:"Duration" default:"518400" require:"true" description:"Set the price of the ask for retrievals"`
	FileSize     string          `json:"FileSize" jsonschema:"FileSize" title:"FileSize" default:"1M" require:"true" description:"File size in bytes (b=512, kB=1000, K=1024, MB=kB*kB, M=K*K, GB=kB*kB*kB, G=K*K*K)"`

	StatelessDeal      bool            `json:"StatelessDeal"  jsonschema:"StatelessDeal"  title:"StatelessDeal" default:"false" require:"true" description:"true离线订单/false在线订单"`
	From               address.Address `json:"From"  jsonschema:"From"  title:"From" require:"false" description:"From"`
	StartEpoch         int64           `json:"StartEpoch"  jsonschema:"StartEpoch"  title:"StartEpoch" default:"-1" require:"false" description:"StartEpoch"`
	FastRetrieval      bool            `json:"FastRetrieval"  jsonschema:"FastRetrieval"  title:"FastRetrieval" default:"true" require:"true" description:"FastRetrieval"`
}

type InitOfflineOrderReturn struct {
	ProposalCid string
	CarFile     string
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*InitOfflineOrderReturn, error) {
	api, closer, err := clientapi.DialIMarketClientRPC(ctx, params.DropletClient.SvcEndpoint.ToMultiAddr(), params.DropletClient.ClientToken, nil)
	if err != nil {
		return nil, fmt.Errorf("dial market client rpc failed")
	}
	defer closer()

	mountPath := "/carfile/"
	filePath := mountPath + params.PieceStore.Name + "/file.txt"
	carFile := filePath + ".car"
	log.Debugln("carFilePath:", carFile)

	tmpl, err := template.New("command").Parse("dd if=/dev/urandom of={{.FilePath}} bs={{.BlockSize}} count=1")
	if err != nil {
		return nil, fmt.Errorf("parase template: %v", err)
	}

	data := map[string]interface{}{
		"FilePath":  filePath,
		"BlockSize": params.FileSize,
	}

	var createFileCmd bytes.Buffer
	err = tmpl.Execute(&createFileCmd, data)
	if err != nil {
		panic(err)
	}

	err = dropletclient.AddPieceStoragge(ctx, k8sEnv, params.DropletClient, params.PieceStore.Name, mountPath)
	if err != nil {
		return nil, err
	}

	pods, err := dropletclient.GetPods(ctx, k8sEnv, params.DropletClient.InstanceName)
	if err != nil {
		return nil, err
	}

	_, err = k8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), "/bin/sh", "-c", createFileCmd.String())
	if err != nil {
		return nil, err
	}

	dataImportReturns, err := GetCidAndPieceSize(ctx, api, filePath, carFile)
	if err != nil {
		return nil, err
	}

	proposalCid, err := StorageDealsInit(ctx, params, api, dataImportReturns)
	if err != nil {
		return nil, err
	}

	return &InitOfflineOrderReturn{
		CarFile:     carFile,
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

	log.Debugln("data:", data)
	log.Debugln("wallet:", wallet)
	log.Debugln("params.MinerAddress:", params.MinerAddress)
	log.Debugln("EpochPrice:", vtypes.BigInt(params.Price))
	log.Debugln("MinBlocksDuration:", uint64(params.Duration))
	log.Debugln("DealStartEpoch:", abi.ChainEpoch(params.StartEpoch))
	log.Debugln("FastRetrieval:", params.FastRetrieval)

	sdParams := &client.DealParams{
		Data:               data,
		Wallet:             wallet,
		Miner:              params.MinerAddress,
		EpochPrice:         vtypes.BigInt(params.Price),
		MinBlocksDuration:  uint64(params.Duration),
		DealStartEpoch:     abi.ChainEpoch(params.StartEpoch),
		FastRetrieval:      params.FastRetrieval,
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

	log.Debugln("DealCid cid: ", encoder.Encode(*proposal))
	return encoder.Encode(*proposal), nil
}
