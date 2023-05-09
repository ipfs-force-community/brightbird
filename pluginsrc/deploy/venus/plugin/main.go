package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/pluginsrc/deploy/venus"
	"github.com/hunjixin/brightbird/types"
)

var Info = venus.PluginInfo

type DepParams struct {
	Params venus.Config `optional:"true"`

	VenusAuthDeploy env.IDeployer `svcname:"VenusAuth"`

	K8sEnv         *env.K8sEnvDeployer
	BootstrapPeers types.BootstrapPeers
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	adminToken, err := depParams.VenusAuthDeploy.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	deployer, err := venus.DeployerFromConfig(depParams.K8sEnv, venus.Config{
		AuthUrl:        depParams.VenusAuthDeploy.SvcEndpoint().ToHttp(),
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
