package main

import (
	"context"
	"fmt"

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
	Name:        "test_webhook",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "just test webhook",
}

type TestCaseParams struct {
	fx.In
	Tester env.IDeployer `json:"-" svcname:"Test"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	fmt.Println("webhook test")
	return env.NewSimpleExec(), nil
}
