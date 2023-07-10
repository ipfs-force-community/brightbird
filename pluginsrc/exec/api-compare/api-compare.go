package main

import (
	"context"

	"github.com/hunjixin/brightbird/types"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/version"
	"github.com/simlecode/api-compare/cmd"
	"github.com/urfave/cli/v2"
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
	VenusURL     string `json:"venusUrl" jsonschema:"venusUrl" title:"Venus Url" require:"true" desctiption:"url for connect venus"`
	VenusToken   string `json:"venusToken" jsonschema:"venusToken" title:"Venus Token" require:"true" desctiption:"token for connect venus"`
	LotusURL     string `json:"lotusUrl" jsonschema:"lotusUrl" title:"Lotus Url" require:"true" desctiption:"url for connect lotus"`
	LotusToken   string `json:"lotusToken" jsonschema:"lotusToken" title:"Lotus Token" require:"true" desctiption:"url to connect lotus"`
	StopHeight   string `json:"stopHeight" jsonschema:"stopHeight" title:"StopHeight" require:"true" desctiption:"check until specific height"`
	EnableEthRPC bool   `json:"enableEthRpc" jsonschema:"enableEthRpc" title:"Enable Eth API" require:"true" default:"true" desctiption:"enable check eth api"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	cliCtx := &cli.Context{
		Context: ctx,
	}
	if params.VenusURL != "" {
		err := cliCtx.Set("venus-url", params.VenusURL)
		if err != nil {
			return err
		}
	}
	if params.VenusToken != "" {
		err := cliCtx.Set("venus-token", params.VenusToken)
		if err != nil {
			return err
		}
	}

	if params.LotusURL != "" {
		err := cliCtx.Set("lotus-url", params.LotusURL)
		if err != nil {
			return err
		}
	}

	if params.LotusToken != "" {
		err := cliCtx.Set("lotus-token", params.LotusToken)
		if err != nil {
			return err
		}
	}

	if params.StopHeight != "" {
		err := cliCtx.Set("stop-height", params.StopHeight)
		if err != nil {
			return err
		}
	}

	if params.EnableEthRPC {
		err := cliCtx.Set("enable-eth-rpc", "true")
		if err != nil {
			return err
		}
	}

	err := cmd.Run(cliCtx)
	if err != nil {
		return err
	}

	return nil
}
