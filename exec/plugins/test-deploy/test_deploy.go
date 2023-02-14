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
	Description: "generate admin token",
}

type TestCaseParams struct {
	fx.In
	MarketClient   env.IMarketClientDeployer
	VenusWallet    env.IVenusWalletDeployer `svcname:"Wallet"`
	VenusWalletNew env.IVenusWalletDeployer `svcname:"WalletNew"`
}

func Exec(ctx context.Context, params TestCaseParams) error {
	fmt.Println("Wallet", params.VenusWallet.SvcEndpoint())
	fmt.Println("WalletNew", params.VenusWalletNew.SvcEndpoint())
	return nil
}
