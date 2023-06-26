package main

import (
	"context"

	venusutils "github.com/hunjixin/brightbird/env/venus_utils"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	dropletclient "github.com/hunjixin/brightbird/pluginsrc/deploy/droplet-client"
)

func main() {
	plugin.SetupPluginFromStdin(dropletclient.PluginInfo, Exec)
}

type DepParams struct {
	Params dropletclient.Config `optional:"true"`

	VenusDep     env.IDeployer `svcname:"Venus" description:"[Deploy]venus"`
	WalletDeploy env.IDeployer `svcname:"VenusWallet" description:"[Deploy]venus-wallet"`
	AuthDeploy   env.IDeployer `svcname:"SophonAuth" description:"[Deploy]sophon-auth"`

	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	adminToken, err := depParams.AuthDeploy.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	pods, err := depParams.WalletDeploy.Pods(ctx)
	if err != nil {
		return nil, err
	}

	walletToken, err := venusutils.ReadWalletToken(ctx, depParams.K8sEnv, pods[0].GetName())
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

	deployer, err := dropletclient.DeployerFromConfig(depParams.K8sEnv, dropletclient.Config{
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
