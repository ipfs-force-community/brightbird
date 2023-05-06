package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	venus_miner "github.com/hunjixin/brightbird/env/impl/venus-miner"
	"github.com/hunjixin/brightbird/types"
)

var Info = venus_miner.PluginInfo

type DepParams struct {
	Params venus_miner.Config `optional:"true"`

	VenusAuth env.IDeployer `svcname:"VenusAuth"`
	Venus     env.IDeployer `svcname:"Venus"`
	Gateway   env.IDeployer `svcname:"VenusGateway"`

	K8sEnv     *env.K8sEnvDeployer
	AdminToken types.AdminToken
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	deployer, err := venus_miner.DeployerFromConfig(depParams.K8sEnv, venus_miner.Config{
		NodeUrl:    depParams.Venus.SvcEndpoint().ToMultiAddr(),
		GatewayUrl: depParams.Gateway.SvcEndpoint().ToMultiAddr(),
		AuthUrl:    depParams.VenusAuth.SvcEndpoint().ToHttp(),
		AuthToken:  string(depParams.AdminToken),
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
