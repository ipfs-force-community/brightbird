package main

import (
	"context"
	"encoding/json"
	"github.com/hunjixin/brightbird/env"
	venus_gateway "github.com/hunjixin/brightbird/env/impl/venus-gateway"
	"github.com/hunjixin/brightbird/types"
)

var Info = venus_gateway.PluginInfo

type DepParams struct {
	Params     json.RawMessage `optional:"true"`
	K8sEnv     *env.K8sEnvDeployer
	VenusAuth  env.IVenusAuthDeployer
	AdminToken types.AdminToken
}

func Exec(ctx context.Context, depParams DepParams) (env.IVenusGatewayDeployer, error) {
	deployer, err := venus_gateway.DeployerFromConfig(depParams.K8sEnv, venus_gateway.Config{
		AuthUrl: depParams.VenusAuth.SvcEndpoint().ToHttp(),
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
