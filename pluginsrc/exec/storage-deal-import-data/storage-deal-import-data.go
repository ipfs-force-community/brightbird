package main

import (
	"context"

	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletmarket "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/pvc"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("storage-deal-import-data")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "storage-deal-import-data",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "手动导入deal数据",
}

type TestCaseParams struct {
	Droplet    dropletmarket.DropletMarketDeployReturn `json:"Droplet" jsonschema:"Droplet" title:"Droplet" description:"droplet return"`
	PieceStore pvc.PvcReturn                           `json:"PieceStore" jsonschema:"PieceStore" title:"PieceStore" require:"true" description:"piece storage"`

	ProposalCid string `json:"ProposalCid"  jsonschema:"ProposalCid"  title:"ProposalCid" require:"true" description:"ProposalCid"`
	CarFile     string `json:"carFile"  jsonschema:"carFile"  title:"carFile" require:"true" description:"carFile"`
	SkipCommP   bool   `json:"skipCommP"  jsonschema:"skipCommP"  title:"skipCommP" require:"true" default:"false" description:"skip calculate the piece-cid"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	client, closer, err := marketapi.DialIMarketRPC(ctx, params.Droplet.SvcEndpoint.ToMultiAddr(), params.Droplet.UserToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	mountPath := "/carfile/"
	err = dropletmarket.AddPieceStoragge(ctx, k8sEnv, params.Droplet, params.PieceStore.Name, mountPath)
	if err != nil {
		return err
	}

	proposalCid, err := cid.Decode(params.ProposalCid)
	if err != nil {
		return err
	}

	log.Debug("proposalCid: ", proposalCid)

	err = client.DealsImportData(ctx, proposalCid, params.CarFile, params.SkipCommP)
	if err != nil {
		return err
	}

	return nil
}
