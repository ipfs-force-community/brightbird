package main

import (
	"context"
	"encoding/json"
	"github.com/hunjixin/brightbird/env"
	venus_wallet "github.com/hunjixin/brightbird/env/impl/venus-wallet"
	"github.com/hunjixin/brightbird/types"
)

var Info = venus_wallet.PluginInfo

type DepParams struct {
	Params     json.RawMessage `optional:"true"`
	K8sEnv     *env.K8sEnvDeployer
	Gateway    env.IVenusGatewayDeployer
	AdminToken types.AdminToken

	types.AnnotateOut
}

func Exec(ctx context.Context, depParams DepParams) (env.IVenusWalletDeployer, error) {
	deployer, err := venus_wallet.DeployerFromConfig(depParams.K8sEnv, venus_wallet.Config{
		GatewayUrl: depParams.Gateway.SvcEndpoint().ToMultiAddr(),
		AuthToken:  string(depParams.AdminToken),
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
