package main

import (
	"context"
	"fmt"

	"github.com/hunjixin/brightbird/types"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/version"
	"github.com/simlecode/api-compare/cmd"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "api_compare",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "api compare",
}

type TestCaseParams struct {
	fx.In
	Params struct {
		VenusURL     string `json:"venus_url"`
		VenusToken   string `json:"venus_token"`
		LotusURL     string `json:"lotus_url"`
		LotusToken   string `json:"lotus_token"`
		StopHeight   string `json:"stop_height"`
		EnableEthRPC string `json:"enable_eth_rpc"`
	} `optional:"true"`
	K8sEnv *env.K8sEnvDeployer `json:"-"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	cliCtx := &cli.Context{
		Context: ctx,
	}
	if params.Params.VenusURL != "" {
		err := cliCtx.Set("venus-url", params.Params.VenusURL)
		if err != nil {
			return nil, err
		}
	}
	if params.Params.VenusToken != "" {
		err := cliCtx.Set("venus-token", params.Params.VenusToken)
		if err != nil {
			return nil, err
		}
	}

	if params.Params.LotusURL != "" {
		err := cliCtx.Set("lotus-url", params.Params.LotusURL)
		if err != nil {
			return nil, err
		}
	}

	if params.Params.LotusToken != "" {
		err := cliCtx.Set("lotus-token", params.Params.LotusToken)
		if err != nil {
			return nil, err
		}
	}

	if params.Params.StopHeight != "" {
		err := cliCtx.Set("stop-height", params.Params.StopHeight)
		if err != nil {
			return nil, err
		}
	}

	if params.Params.EnableEthRPC != "" {
		err := cliCtx.Set("enable-eth-rpc", params.Params.EnableEthRPC)
		if err != nil {
			return nil, err
		}
	}

	err := cmd.Run(cliCtx)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return env.NewSimpleExec(), nil
}
