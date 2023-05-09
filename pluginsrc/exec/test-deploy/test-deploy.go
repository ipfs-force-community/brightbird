package main

import (
	"context"
	"fmt"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "test_deploy",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "generate admin token",
}

type TestCaseParams struct {
	fx.In
	K8sEnv         *env.K8sEnvDeployer `json:"-"`
	MarketClient   env.IDeployer       `json:"-" svcname:"MarketClient"`
	VenusWallet    env.IDeployer       `json:"-" svcname:"VenusWallet"`
	VenusWalletNew env.IDeployer       `json:"-" svcname:"VenusWalletNew"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	//restart pod
	pods, err := params.VenusWallet.Pods(ctx)
	if err != nil {
		return nil, err
	}

	err = params.K8sEnv.StopPods(ctx, pods)
	if err != nil {
		return nil, err
	}

	fmt.Println("Wallet", params.VenusWallet.SvcEndpoint())
	fmt.Println("WalletNew", params.VenusWalletNew.SvcEndpoint())
	return env.NewSimpleExec(), nil
}
