package main

import (
	"context"
	"errors"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	venuswallet "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-wallet"
	"github.com/hunjixin/brightbird/types"
)

func main() {
	plugin.SetupPluginFromStdin(venuswallet.PluginInfo, Exec)
}

type DepParams struct {
	venuswallet.Config
	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (*venuswallet.VenusWalletDeployParams, error) {
	if len(depParams.GatewayUrl) != 0 && len(depParams.UserToken) == 0 {
		return nil, errors.New("gateway have value but not set token value")
	}

	if len(depParams.GatewayUrl) == 0 && len(depParams.UserToken) == 0 {
		return nil, errors.New("token have value but not set gateway url")
	}
	var deployer *venuswallet.VenusWalletDeployer
	var err error
	if len(depParams.UserToken) == 0 {
		deployer, err = venuswallet.DeployerFromConfig(depParams.K8sEnv, venuswallet.Config{
			BaseConfig: depParams.BaseConfig,
			GatewayUrl: types.Endpoint(depParams.GatewayUrl).ToMultiAddr(),
			UserToken:  depParams.UserToken,
		})
	} else {
		deployer, err = venuswallet.DeployerFromConfig(depParams.K8sEnv, venuswallet.Config{
			BaseConfig: depParams.BaseConfig,
		})
	}
	if err != nil {
		return nil, err
	}
	return deployer.Deploy(ctx)
}
