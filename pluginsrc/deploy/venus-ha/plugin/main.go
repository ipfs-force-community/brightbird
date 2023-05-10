package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/types"
	venus_ha "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-ha"
)

var Info = venus_ha.PluginInfo

type DepParams struct {
	Params venus_ha.Config `optional:"true"`

	VenusAuth env.IDeployer `svcname:"VenusAuth"`

	K8sEnv         *env.K8sEnvDeployer
	BootstrapPeers types.BootstrapPeers
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	adminToken, err := depParams.VenusAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	deployer, err := venus_ha.DeployerFromConfig(depParams.K8sEnv, venus_ha.Config{
		AuthUrl:        depParams.VenusAuth.SvcEndpoint().ToHttp(),
		AdminToken:     adminToken.(string),
		BootstrapPeers: depParams.BootstrapPeers,
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
