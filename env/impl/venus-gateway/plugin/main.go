package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	venus_gateway "github.com/hunjixin/brightbird/env/impl/venus-gateway"
	"github.com/hunjixin/brightbird/types"
)

var Info = venus_gateway.PluginInfo

type DepParams struct {
	Params venus_gateway.Config `optional:"true"`

	VenusAuth env.IDeployer `svcname:"VenusAuth"`

	K8sEnv     *env.K8sEnvDeployer
	AdminToken types.AdminToken
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	deployer, err := venus_gateway.DeployerFromConfig(depParams.K8sEnv, venus_gateway.Config{
		AuthUrl:    depParams.VenusAuth.SvcEndpoint().ToHttp(),
		AdminToken: depParams.AdminToken,
	}, depParams.Params)
	if err != nil {
		return nil, err
	}

	err = deployer.Deploy(ctx)
	if err != nil {
		return nil, err
	}
	return deployer, nil
}
