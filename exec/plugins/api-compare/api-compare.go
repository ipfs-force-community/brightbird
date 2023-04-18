package api_compare

import (
	"context"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"github.com/simlecode/api-compare/cmd"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "venus-auth-test",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "compare venus and lotus api",
}

type TestCaseParams struct {
	fx.In
	Params struct {
		venusUrl   string `json:"venus_url"`
		venusToken string `json:"venus_token"`
		lotusUrl   string `json:"lotus_url"`
		lotusToken string `json:"lotus_token"`
	} `optional:"true"`
	K8sEnv *env.K8sEnvDeployer `json:"-"`
}

func Exec(ctx context.Context, params TestCaseParams) error {
	cliCtx := &cli.Context{
		Context: ctx,
	}
	cliCtx.Set("venus-url", params.Params.venusUrl)
	cliCtx.Set("venus-token", params.Params.venusToken)
	cliCtx.Set("lotus-url", params.Params.lotusUrl)
	cliCtx.Set("lotus-token", params.Params.lotusToken)

	return cmd.Run(cliCtx)
}
