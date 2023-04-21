package main

import (
	"context"
	"fmt"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"github.com/simlecode/api-compare/cmd"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "api_compare",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "api compare",
}

type TestCaseParams struct {
	fx.In
	Params struct {
		venusUrl     string `json:"venus_url"`
		venusToken   string `json:"venus_token"`
		lotusUrl     string `json:"lotus_url"`
		lotusToken   string `json:"lotus_token"`
		stopHeight   string `json:"stop_height"`
		enableEthRpc string `json:"enable_eth_rpc"`
	} `optional:"true"`
	K8sEnv *env.K8sEnvDeployer `json:"-"`
}

func Exec(ctx context.Context, params TestCaseParams) error {
	cliCtx := &cli.Context{
		Context: ctx,
	}
	if params.Params.venusUrl != "" {
		cliCtx.Set("venus-url", params.Params.venusUrl)
	}
	if params.Params.venusToken != "" {
		cliCtx.Set("venus-token", params.Params.venusToken)
	}

	if params.Params.lotusUrl != "" {
		cliCtx.Set("lotus-url", params.Params.lotusUrl)
	}

	if params.Params.lotusToken != "" {
		cliCtx.Set("lotus-token", params.Params.lotusToken)
	}

	if params.Params.stopHeight != "" {
		cliCtx.Set("stop-height", params.Params.stopHeight)
	}

	if params.Params.enableEthRpc != "" {
		cliCtx.Set("enable-eth-rpc", params.Params.enableEthRpc)
	}

	err := cmd.Run(cliCtx)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
