package main

import (
	"context"
	"errors"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	sophongateway "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-gateway"
	venuswallet "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-wallet"
)

func main() {
	plugin.SetupPluginFromStdin(venuswallet.PluginInfo, Exec)
}

type DepParams struct {
	venuswallet.Config
	Gateway *sophongateway.SophonGatewayReturn `json:"SophonGateway" description:"gateway return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*venuswallet.VenusWalletReturn, error) {
	if depParams.Gateway != nil && len(depParams.UserToken) == 0 {
		return nil, errors.New("gateway have value but not set token value")
	}

	if depParams.Gateway == nil && len(depParams.UserToken) == 0 {
		return nil, errors.New("token have value but not set gateway url")
	}

	if len(depParams.UserToken) > 0 {
		return venuswallet.DeployFromConfig(ctx, k8sEnv, venuswallet.Config{
			BaseConfig: depParams.BaseConfig,
			VConfig: venuswallet.VConfig{
				GatewayUrl: depParams.Gateway.SvcEndpoint.ToMultiAddr(),
				UserToken:  depParams.UserToken,
			},
		})
	}
	return venuswallet.DeployFromConfig(ctx, k8sEnv, venuswallet.Config{
		BaseConfig: depParams.BaseConfig,
	})
}
