package main

import (
	"context"
	"fmt"

	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	dropletmarket "github.com/hunjixin/brightbird/pluginsrc/deploy/droplet-market"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "market_listen",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "market listen",
}

type TestCaseParams struct {
	Auth          sophonauth.SophonAuthDeployReturn       `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	DropletMarket dropletmarket.DropletMarketDeployReturn `json:"DropletMarket" jsonschema:"DropletMarket" title:"DropletMarket" description:"droplet market return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	listenAddress, err := marketListen(ctx, params)
	if err != nil {
		fmt.Printf("market net listen err: %v\n", err)
		return err
	}
	fmt.Printf("market net listen is: %v\n", listenAddress)
	return nil
}

func marketListen(ctx context.Context, params TestCaseParams) (string, error) {
	client, closer, err := marketapi.NewIMarketRPC(ctx, params.DropletMarket.SvcEndpoint.ToMultiAddr(), nil)
	if err != nil {
		return "", err
	}
	defer closer()

	addrs, err := client.NetAddrsListen(ctx)
	if err != nil && addrs.Addrs != nil {
		return addrs.String(), nil
	}

	for _, peer := range addrs.Addrs {
		fmt.Printf("%s/p2p/%s\n", peer, addrs.ID)
	}
	return "", err
}
