package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"

	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "venus-auth-test",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "just test venus-auth",
}

type TestCaseParams struct {
	fx.In
	VenusWallet env.IDeployer `json:"-" svcname:"VenusWallet"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	return env.NewSimpleExec(), nil
}
