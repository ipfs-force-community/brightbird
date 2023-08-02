package main

import (
	"context"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	sophongateway "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-gateway"
	venuswalletpro "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus-wallet-pro"
)

func main() {
	plugin.SetupPluginFromStdin(venuswalletpro.PluginInfo, Exec)
}

type DepParams struct {
	venuswalletpro.Config
	Gateway sophongateway.SophonGatewayReturn `json:"SophonGateway"  jsonschema:"SophonGateway"  title:"SophonGateway" require:"true" description:"gateway deploy return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*venuswalletpro.VenusWalletProDeployReturn, error) {
	return venuswalletpro.DeployFromConfig(ctx, k8sEnv, venuswalletpro.Config{
		BaseConfig: depParams.BaseConfig,
		VConfig: venuswalletpro.VConfig{
			GatewayUrl: depParams.Gateway.SvcEndpoint.ToMultiAddr(),
			UserToken:  depParams.UserToken,
			Replicas:   1,
		},
	})
}
