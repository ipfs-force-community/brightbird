package main

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/venus-shared/actors/policy"
	clientapi "github.com/filecoin-project/venus/venus-shared/api/market/client"
	"github.com/filecoin-project/venus/venus-shared/types/market/client"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletclient "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-client"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("init-offline-deal-v2")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "init-offline-deal-v2",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "droplet-client发起v2版本离线订单",
}

type TestCaseParams struct {
	DropletClient dropletclient.DropletClientDeployReturn `json:"DropClient" jsonschema:"DropClient" title:"DropletClient" description:"droplet client return"`

	File    string `json:"File" jsonschema:"File" title:"File"`
	CarFile string `json:"CarFile" jsonschema:"CarFile" title:"CarFile"`

	MinerAddress address.Address `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true"`
	Duration     int64           `json:"Duration"  jsonschema:"Duration"  title:"Duration" default:"518400" require:"true" description:"Set the price of the ask for retrievals"`
	From         address.Address `json:"From"  jsonschema:"From"  title:"From" require:"false" description:"From"`
}

type InitOfflineDealReturn struct {
	ProposalCid string
	DealUUID    string
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*InitOfflineDealReturn, error) {
	api, closer, err := clientapi.DialIMarketClientRPC(ctx, params.DropletClient.SvcEndpoint.ToMultiAddr(), params.DropletClient.ClientToken, nil)
	if err != nil {
		return nil, fmt.Errorf("dial market client rpc failed")
	}
	defer closer()

	initData, err := getCidAndPieceSize(ctx, api, params.File, params.CarFile)
	if err != nil {
		return nil, err
	}

	wallet := params.From
	if wallet == address.Undef {
		wallet, err = api.DefaultAddress(ctx)
		if err != nil {
			return nil, err
		}
	}

	duration := params.Duration
	MinDealDuration, MaxDealDuration := policy.DealDurationBounds(0)
	if abi.ChainEpoch(duration) < MinDealDuration {
		return nil, fmt.Errorf("minimum deal duration is %d blocks", MinDealDuration)
	}
	if abi.ChainEpoch(duration) > MaxDealDuration {
		return nil, fmt.Errorf("maximum deal duration is %d blocks", MaxDealDuration)
	}

	tmpl, err := template.New("command").Parse("./droplet-client storage deals init-v2 --from={{.From}} --manual-stateless-deal --piece-cid={{.PieceCid}} --piece-size={{.PieceSize}} {{.DataCid}} {{.Miner}} 0 {{.Duration}}")
	if err != nil {
		return nil, fmt.Errorf("parse template: %v", err)
	}

	data := map[string]interface{}{
		"From":      wallet,
		"PieceCid":  initData.Cid,
		"PieceSize": initData.PieceSize,
		"DataCid":   initData.Root,
		"Miner":     params.MinerAddress,
		"Duration":  duration,
	}

	var initDealCmd bytes.Buffer
	err = tmpl.Execute(&initDealCmd, data)
	if err != nil {
		return nil, fmt.Errorf("execute template: %v", err)
	}

	pods, err := dropletclient.GetPods(ctx, k8sEnv, params.DropletClient.InstanceName)
	if err != nil {
		return nil, err
	}

	res, err := k8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), "/bin/sh", "-c", initDealCmd.String())
	if err != nil {
		return nil, err
	}

	fmt.Printf("init deal result: %s\n", res)

	var dealUUID, proposalCid string
	for i, line := range bytes.Split(res, []byte("\n")) {
		if i == 0 {
			dealUUID = string(line[len("deal uuid:  "):])
		}
		if i == 1 {
			proposalCid = string(line[len("proposal cid:  "):])
		}
	}
	fmt.Println("deal uuid: ", dealUUID, " proposal cid: ", proposalCid)

	return &InitOfflineDealReturn{
		ProposalCid: proposalCid,
		DealUUID:    dealUUID,
	}, nil
}

type DataImportReturn struct {
	Root      cid.Cid
	Cid       cid.Cid
	PieceSize abi.UnpaddedPieceSize
}

func getCidAndPieceSize(ctx context.Context, api clientapi.IMarketClient, file, carFile string) (*DataImportReturn, error) {
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

	log.Infoln("Root is: ", c.Root)
	log.Infoln("Cid is: ", ret.Root)
	log.Infoln("PieceSize is: ", ret.Size)

	return &DataImportReturn{
		Root:      c.Root,
		Cid:       ret.Root,
		PieceSize: ret.Size,
	}, nil
}
