package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
)

func main() {
	plugin.SetupPluginFromStdin(sophonauth.PluginInfo, Exec)
}

type DepParams struct {
	Params sophonauth.Config `optional:"true"`
	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	deployer, err := sophonauth.DeployerFromConfig(depParams.K8sEnv, sophonauth.Config{
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
