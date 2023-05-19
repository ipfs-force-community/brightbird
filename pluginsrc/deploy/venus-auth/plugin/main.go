package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	venus_auth "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-auth"
)

func main() {
	plugin.SetupPluginFromStdin(venus_auth.PluginInfo, Exec)
}

type DepParams struct {
	Params venus_auth.Config `optional:"true"`
	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	deployer, err := venus_auth.DeployerFromConfig(depParams.K8sEnv, venus_auth.Config{
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
