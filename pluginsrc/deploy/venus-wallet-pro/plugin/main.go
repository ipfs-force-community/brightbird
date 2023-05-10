package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	venus_wallet_pro "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-wallet-pro"
)

func main() {
	plugin.SetupPluginFromStdin(venus_wallet_pro.PluginInfo, Exec)
}

type DepParams struct {
	Params venus_wallet_pro.Config `optional:"true"`
	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	deployer, err := venus_wallet_pro.DeployerFromConfig(depParams.K8sEnv, venus_wallet_pro.Config{
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
