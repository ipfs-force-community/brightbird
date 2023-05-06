package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	market_client "github.com/hunjixin/brightbird/env/impl/market-client"
	"github.com/hunjixin/brightbird/types"
)

var Info = market_client.PluginInfo

type DepParams struct {
	Params market_client.Config `optional:"true"`

	VenusDep     env.IDeployer `svcname:"Venus"`
	WalletDeploy env.IDeployer `svcname:VenusWallet`

	K8sEnv     *env.K8sEnvDeployer
	AdminToken types.AdminToken
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	pods, err := depParams.WalletDeploy.Pods(ctx)
	if err != nil {
		return nil, err
	}
	walletToken, err := env.ReadWalletToken(ctx, depParams.K8sEnv, pods[0].GetName())
	if err != nil {
		return nil, err
	}
	deployer, err := market_client.DeployerFromConfig(depParams.K8sEnv, market_client.Config{
		NodeUrl:     depParams.VenusDep.SvcEndpoint().ToMultiAddr(),
		NodeToken:   string(depParams.AdminToken),
		WalletUrl:   depParams.WalletDeploy.SvcEndpoint().ToMultiAddr(),
		WalletToken: walletToken,
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
