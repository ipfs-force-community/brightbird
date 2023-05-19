package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	venus_worker "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-worker"
)

func main() {
	plugin.SetupPluginFromStdin(venus_worker.PluginInfo, Exec)
}

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

	sectorManagerEndpoint, err := depParams.SectorManager.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	deployer, err := venus_worker.DeployerFromConfig(depParams.K8sEnv, venus_worker.Config{
		VenusSectorManagerUrl: sectorManagerEndpoint.ToHttp(),
		AuthToken:             adminToken.MustString(),
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
