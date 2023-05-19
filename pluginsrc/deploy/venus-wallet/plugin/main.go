package main

import (
	"context"
	"errors"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	venus_wallet "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-wallet"
	"github.com/hunjixin/brightbird/types"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("venus-wallet-dep")

func main() {
	plugin.SetupPluginFromStdin(venus_wallet.PluginInfo, Exec)
}

type DepParams struct {
	Params venus_wallet.Config `optional:"true"`

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

		deployer, err = venus_wallet.DeployerFromConfig(depParams.K8sEnv, venus_wallet.Config{
			GatewayUrl: gatewayEndpoint.ToMultiAddr(),
			UserToken:  userToken.MustString(),
		}, depParams.Params)
	} else {
		deployer, err = venus_wallet.DeployerFromConfig(depParams.K8sEnv, venus_wallet.Config{}, depParams.Params)
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
