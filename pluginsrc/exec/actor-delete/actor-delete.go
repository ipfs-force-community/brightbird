package main

import (
	"context"

	"github.com/filecoin-project/go-address"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletmarket "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("actor-delete")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "actor-delete",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "删除miner",
}

type TestCaseParams struct {
	Droplet      dropletmarket.DropletMarketDeployReturn `json:"Droplet" jsonschema:"Droplet" title:"Droplet" description:"droplet return"`
	MinerAddress address.Address                         `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true" `
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	client, closer, err := marketapi.DialIMarketRPC(ctx, params.Droplet.SvcEndpoint.ToMultiAddr(), params.Droplet.UserToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	err = client.ActorDelete(ctx, params.MinerAddress)
	if err != nil {
		return err
	}

	err = actorList(ctx, params, client)
	if err != nil {
		log.Errorln("list actor err %w", err)
		return err
	}

	log.Debugf("delete miner %s success\n", params.MinerAddress)
	return nil
}

func actorList(ctx context.Context, params TestCaseParams, api marketapi.IMarket) error {
	miners, err := api.ActorList(ctx)
	if err != nil {
		return err
	}
	for _, miner := range miners {
		if miner.Addr == params.MinerAddress {
			return err
		}
	}

	return nil
}
