package main

import (
	"context"

	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	"github.com/ipfs/go-cid"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletmarket "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

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
	Droplet dropletmarket.DropletMarketDeployReturn `json:"Droplet" jsonschema:"Droplet" title:"Droplet" description:"droplet return"`

	DealPropCid *cid.Cid `json:"DealPropCid"  jsonschema:"DealPropCid"  title:"DealPropCid" require:"true" description:"DealPropCid"`
	CarFile     string   `json:"carFile"  jsonschema:"carFile"  title:"carFile" require:"true" description:"carFile"`
	SkipCommP   bool     `json:"skipCommP"  jsonschema:"skipCommP"  title:"skipCommP" default:"false" require:"true" description:"skip calculate the piece-cid"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	client, closer, err := marketapi.DialIMarketRPC(ctx, params.Droplet.SvcEndpoint.ToMultiAddr(), params.Droplet.UserToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	err = client.DealsImportData(ctx, *params.DealPropCid, params.CarFile, params.SkipCommP)
	if err != nil {
		return err
	}

	return nil
}
