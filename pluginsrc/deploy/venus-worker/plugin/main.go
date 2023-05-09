package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	venus_worker "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-worker"
)

var Info = venus_worker.PluginInfo

type DepParams struct {
	Params venus_worker.Config `optional:"true"`

	VenusAuth     env.IDeployer `svcname:"VenusAuth"`
	SectorManager env.IDeployer `svcname:"VenusSectorManager"`
	WalletDeploy  env.IDeployer `svcname:"VenusWallet"`

	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	adminToken, err := depParams.VenusAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	deployer, err := venus_worker.DeployerFromConfig(depParams.K8sEnv, venus_worker.Config{
		VenusSectorManagerUrl: depParams.SectorManager.SvcEndpoint().ToHttp(),
		AuthToken:             adminToken.(string),
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
