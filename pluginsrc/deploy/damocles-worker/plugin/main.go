package main

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	damoclesworker "github.com/hunjixin/brightbird/pluginsrc/deploy/damocles-worker"
)

func main() {
	plugin.SetupPluginFromStdin(damoclesworker.PluginInfo, Exec)
}

type DepParams struct {
	Params damoclesworker.Config `optional:"true"`

	SophonAuth    env.IDeployer `svcname:"SophonAuth"`
	SectorManager env.IDeployer `svcname:"DamoclesManager"`
	WalletDeploy  env.IDeployer `svcname:"VenusWallet"`

	K8sEnv *env.K8sEnvDeployer

	Miner env.IExec `svcname:"MinerInfo"`
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	adminToken, err := depParams.SophonAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	sectorManagerEndpoint, err := depParams.SectorManager.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	minerP, err := depParams.Miner.Param("Miner")
	if err != nil {
		return nil, err
	}
	minerAddr, err := env.UnmarshalJSON[address.Address](minerP.Raw())
	if err != nil {
		return nil, err
	}

	deployer, err := damoclesworker.DeployerFromConfig(depParams.K8sEnv, damoclesworker.Config{
		DamoclesManagerURL: sectorManagerEndpoint.ToHTTP(),
		AuthToken:          adminToken.MustString(),
		MinerAddress:       minerAddr,
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
