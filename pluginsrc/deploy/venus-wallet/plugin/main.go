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
	Params venuswallet.Config `optional:"true"`

	Gateway     env.IDeployer `optional:"true" svcname:"VenusGateway"`
	CreateToken env.IExec     `optional:"true" svcname:"Token"`

	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	if depParams.Gateway != nil && depParams.CreateToken == nil {
		return nil, errors.New("gateway have value but not set token value")
	}

	if depParams.Gateway == nil && depParams.CreateToken != nil {
		return nil, errors.New("token have value but not set gateway url")
	}
	var deployer env.IDeployer
	var err error
	if depParams.CreateToken != nil {
		var userToken env.Params
		userToken, err = depParams.CreateToken.Param("Token")
		if err != nil && err != env.ErrParamsNotFound {
			return nil, err
		}

		var gatewayEndpoint types.Endpoint
		gatewayEndpoint, err = depParams.Gateway.SvcEndpoint()
		if err != nil {
			return nil, err
		}

		deployer, err = venuswallet.DeployerFromConfig(depParams.K8sEnv, venuswallet.Config{
			GatewayUrl: gatewayEndpoint.ToMultiAddr(),
			UserToken:  userToken.MustString(),
		}, depParams.Params)
	} else {
		deployer, err = venuswallet.DeployerFromConfig(depParams.K8sEnv, venuswallet.Config{}, depParams.Params)
	}
	if err != nil {
		return nil, err
	}
	err = deployer.Deploy(ctx)
	if err != nil {
		return nil, err
	}
	return deployer, nil
}
