package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	venus_worker "github.com/hunjixin/brightbird/env/impl/venus-worker"
	"github.com/hunjixin/brightbird/types"
)

var Info = venus_worker.PluginInfo

type DepParams struct {
	Params venus_worker.Config `optional:"true"`

	SectorManager env.IDeployer `svcname:"VenusSectorManager"`
	WalletDeploy  env.IDeployer `svcname:VenusWallet`

	K8sEnv     *env.K8sEnvDeployer
	AdminToken types.AdminToken
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	deployer, err := venus_worker.DeployerFromConfig(depParams.K8sEnv, venus_worker.Config{
		VenusSectorManagerUrl: depParams.SectorManager.SvcEndpoint().ToHttp(),
		AuthToken:             string(depParams.AdminToken),
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
