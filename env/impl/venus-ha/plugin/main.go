package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	venus_ha "github.com/hunjixin/brightbird/env/impl/venus-ha"
	"github.com/hunjixin/brightbird/types"
)

var Info = venus_ha.PluginInfo

type DepParams struct {
	Params    venus_ha.Config `optional:"true"`
	K8sEnv    *env.K8sEnvDeployer
	VenusAuth env.IVenusAuthDeployer

	AdminToken     types.AdminToken
	BootstrapPeers types.BootstrapPeers
}

func Exec(ctx context.Context, depParams DepParams) (env.IVenusDeployer, error) {
	deployer, err := venus_ha.DeployerFromConfig(depParams.K8sEnv, venus_ha.Config{
		AuthUrl:        depParams.VenusAuth.SvcEndpoint().ToHttp(),
		AdminToken:     depParams.AdminToken,
		BootstrapPeers: depParams.BootstrapPeers,
	}, depParams.Params)
	if err != nil {
		return nil, err
	}

	err = deployer.Deploy(ctx)
	if err != nil {
		return nil, err
	}
	err = env.SyncWait(ctx, depParams.K8sEnv, deployer, string(depParams.AdminToken))
	if err != nil {
		return nil, err
	}
	return deployer, nil
}
