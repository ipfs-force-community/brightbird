package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	venuswalletpro "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-wallet-pro"
)

func main() {
	plugin.SetupPluginFromStdin(venuswalletpro.PluginInfo, Exec)
}

type DepParams struct {
	Params venuswalletpro.Config `optional:"true"`
	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	deployer, err := venuswalletpro.DeployerFromConfig(depParams.K8sEnv, venuswalletpro.Config{
		Replicas: 1,
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
