package main

import (
	"context"
	"errors"

	"github.com/hunjixin/brightbird/env"
	venus_wallet "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-wallet"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("venus-wallet-dep")
var Info = venus_wallet.PluginInfo

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
		var userToken interface{}
		userToken, err = depParams.CreateToken.Param("Token")
		if err != nil && err != env.ErrParamsNotFound {
			return nil, err
		}

		deployer, err = venus_wallet.DeployerFromConfig(depParams.K8sEnv, venus_wallet.Config{
			GatewayUrl: depParams.Gateway.SvcEndpoint().ToMultiAddr(),
			UserToken:  userToken.(string),
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
