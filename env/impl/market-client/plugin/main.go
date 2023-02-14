package main

import (
	"context"
	"encoding/json"
	"github.com/hunjixin/brightbird/env"
	market_client "github.com/hunjixin/brightbird/env/impl/market-client"
	"github.com/hunjixin/brightbird/types"
)

var Info = market_client.PluginInfo

type DepParams struct {
	Params       json.RawMessage `optional:"true"`
	K8sEnv       *env.K8sEnvDeployer
	VenusDep     env.IVenusDeployer
	WalletDeploy env.IVenusWalletDeployer `svcname:"Wallet"`
	AdminToken   types.AdminToken

	types.AnnotateOut
}

func Exec(ctx context.Context, depParams DepParams) (env.IMarketClientDeployer, error) {
	walletToken, err := env.ReadWalletToken(ctx, depParams.K8sEnv, depParams.WalletDeploy.Pods()[0].GetName())
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
