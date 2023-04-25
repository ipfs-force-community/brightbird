package main

import (
	"context"
	"fmt"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "actor-set-addrs",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "actor set-addrs",
}

type TestCaseParams struct {
	fx.In
	AdminToken  types.AdminToken
	K8sEnv      *env.K8sEnvDeployer      `json:"-"`
	VenusAuth   env.IVenusAuthDeployer   `json:"-"`
	VenusMarket env.IVenusMarketDeployer `json:"-"`
}

func Exec(ctx context.Context, params TestCaseParams) error {

	listenAddress, err := marketListen(ctx, params)
	if err != nil {
		fmt.Printf("market net listen err: %v\n", err)
		return err
	}
	fmt.Printf("market net listen is: %v\n", listenAddress)
	return nil
}

func marketListen(ctx context.Context, params TestCaseParams) (string, error) {
	endpoint := params.VenusMarket.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, params.VenusMarket.Pods()[0].GetName(), int(params.VenusMarket.Svc().Spec.Ports[0].Port))
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
