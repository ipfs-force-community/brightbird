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

var log = logging.Logger("actor-list")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "actor-list",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "从droplet查询miner列表",
}

type TestCaseParams struct {
	Droplet dropletmarket.DropletMarketDeployReturn `json:"Droplet" jsonschema:"Droplet" title:"Droplet" description:"droplet return"`
}

type DropletActorListReturn struct {
	ActorList []ActorInfo `json:"actorList" jsonschema:"actorList" title:"actorList" require:"true" description:"actor list"`
}

type ActorInfo struct {
	MinerAddress address.Address `json:"minerAddress" jsonschema:"minerAddress" title:"minerAddress" require:"true" description:"miner address"`
	Account      string          `json:"account" jsonschema:"account" title:"account" require:"true" description:"account"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*DropletActorListReturn, error) {
	client, closer, err := marketapi.DialIMarketRPC(ctx, params.Droplet.SvcEndpoint.ToMultiAddr(), params.Droplet.UserToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	miners, err := client.ActorList(ctx)
	if err != nil {
		return nil, err
	}

	dropletActorListReturn := &DropletActorListReturn{}

	for _, miner := range miners {
		log.Debugf("%s\t%s\n", miner.Addr.String(), miner.Account)
		actorInfo := &ActorInfo{
			MinerAddress: miner.Addr,
			Account:      miner.Account,
		}
		dropletActorListReturn.ActorList = append(dropletActorListReturn.ActorList, *actorInfo)
	}

	return dropletActorListReturn, nil
}
