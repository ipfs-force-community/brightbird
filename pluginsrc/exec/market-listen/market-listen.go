package main

import (
	"context"
	"fmt"

	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "market_listen",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "market listen",
}

type TestCaseParams struct {
	fx.In
	K8sEnv      *env.K8sEnvDeployer `json:"-"`
	VenusAuth   env.IDeployer       `json:"-" svcname:"VenusAuth"`
	VenusMarket env.IDeployer       `json:"-" svcname:"VenusMarket"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {

	listenAddress, err := marketListen(ctx, params)
	if err != nil {
		fmt.Printf("market net listen err: %v\n", err)
		return nil, err
	}
	fmt.Printf("market net listen is: %v\n", listenAddress)
	return env.NewSimpleExec(), nil
}

func marketListen(ctx context.Context, params TestCaseParams) (string, error) {
	endpoint := params.VenusMarket.SvcEndpoint()
	if env.Debug {
		pods, err := params.VenusMarket.Pods(ctx)
		if err != nil {
			return "", err
		}

		svc, err := params.VenusMarket.Svc(ctx)
		if err != nil {
			return "", err
		}
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return "", err
		}
	}
	client, closer, err := marketapi.NewIMarketRPC(ctx, endpoint.ToHttp(), nil)
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
