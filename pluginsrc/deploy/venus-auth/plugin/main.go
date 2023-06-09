package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	venusauth "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-auth"
)

func main() {
	plugin.SetupPluginFromStdin(venusauth.PluginInfo, Exec)
}

type DepParams struct {
	Params venusauth.Config `optional:"true"`
	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	deployer, err := venusauth.DeployerFromConfig(depParams.K8sEnv, venusauth.Config{
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
