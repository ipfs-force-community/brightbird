package main

import (
	"context"

	"github.com/filecoin-project/go-address"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	mkTypes "github.com/filecoin-project/venus/venus-shared/types/market"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletmarket "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("actor-list")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "actor-upsert",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "更新或者新增miner",
}

type TestCaseParams struct {
	Droplet      dropletmarket.DropletMarketDeployReturn `json:"Droplet" jsonschema:"Droplet" title:"Droplet" description:"droplet return"`
	MinerAddress address.Address                         `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true"`
	Account      string                                  `json:"account"  jsonschema:"account" title:"account" require:"false" description:"create username"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	client, closer, err := marketapi.DialIMarketRPC(ctx, params.Droplet.SvcEndpoint.ToMultiAddr(), params.Droplet.UserToken, nil)
	if err != nil {
		return err
	}
	defer closer()
	err = actorUpsert(ctx, params, client)
	if err != nil {
		log.Errorln("upsert actor err %v", err)
		return err
	}

	err = actorList(ctx, params, client)
	if err != nil {
		log.Errorln("list actor err %w", err)
		return err
	}

	return nil
}

func actorUpsert(ctx context.Context, params TestCaseParams, client marketapi.IMarket) error {
	bAdd, err := client.ActorUpsert(ctx, mkTypes.User{Addr: params.MinerAddress, Account: params.Account})
	if err != nil {
		return err
	}

	opr := "Add"
	if !bAdd {
		opr = "Update"
	}

	log.Debugln("%s miner %s success\n", opr, params.MinerAddress)

	return nil
}

func actorList(ctx context.Context, params TestCaseParams, client marketapi.IMarket) error {
	miners, err := client.ActorList(ctx)
	if err != nil {
		return err
	}
	for _, miner := range miners {
		if miner.Addr == params.MinerAddress {
			return nil
		}
	}

	return err
}
