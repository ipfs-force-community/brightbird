package main

import (
	"context"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "venus-auth-test",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "just test venus-auth",
}

type TestCaseParams struct {
	fx.In
	VenusWallet env.IVenusAuthDeployer `json:"-"`
}

func Exec(ctx context.Context, params TestCaseParams) error {
	return nil
}
