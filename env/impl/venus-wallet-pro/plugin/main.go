package main

import (
	"context"
	"github.com/hunjixin/brightbird/env"
	venus_wallet_pro "github.com/hunjixin/brightbird/env/impl/venus-wallet-pro"
)

var Info = venus_wallet_pro.PluginInfo

type DepParams struct {
	Params venus_wallet_pro.Config `optional:"true"`
	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.IVenusWalletProDeployer, error) {

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
