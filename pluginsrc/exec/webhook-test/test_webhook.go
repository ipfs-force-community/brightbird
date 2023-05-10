package main

import (
	"context"
	"fmt"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "test_webhook",
	Version:     version.Version(),
	Category:    types.TestExec,
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
