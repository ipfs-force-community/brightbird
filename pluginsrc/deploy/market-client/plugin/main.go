package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	marketclient "github.com/hunjixin/brightbird/pluginsrc/deploy/market-client"
)

func main() {
	plugin.SetupPluginFromStdin(marketclient.PluginInfo, Exec)
}

type DepParams struct {
	Params marketclient.Config `optional:"true"`

	VenusDep        env.IDeployer `svcname:"Venus"`
	WalletDeploy    env.IDeployer `svcname:"VenusWallet"`
	VenusAuthDeploy env.IDeployer `svcname:"VenusAuth"`

	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	adminToken, err := depParams.VenusAuthDeploy.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	pods, err := depParams.WalletDeploy.Pods(ctx)
	if err != nil {
		return nil, err
	}

	walletToken, err := env.ReadWalletToken(ctx, depParams.K8sEnv, pods[0].GetName())
	if err != nil {
		return nil, err
	}

	venusEndpoin, err := depParams.VenusDep.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	walletPoint, err := depParams.WalletDeploy.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	deployer, err := marketclient.DeployerFromConfig(depParams.K8sEnv, marketclient.Config{
		NodeUrl:     venusEndpoin.ToMultiAddr(),
		NodeToken:   adminToken.MustString(),
		WalletUrl:   walletPoint.ToMultiAddr(),
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
