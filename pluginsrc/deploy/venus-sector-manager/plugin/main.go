package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	venussectormanager "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-sector-manager"
)

func main() {
	plugin.SetupPluginFromStdin(venussectormanager.PluginInfo, Exec)
}

type DepParams struct {
	Params venussectormanager.Config `optional:"true"`

	VenusAuth    env.IDeployer `svcname:"VenusAuth"`
	Venus        env.IDeployer `svcname:"Venus"`
	Message      env.IDeployer `svcname:"VenusMessager"`
	Gateway      env.IDeployer `svcname:"VenusGateway"`
	WalletDeploy env.IDeployer `svcname:"VenusWallet"`

	K8sEnv *env.K8sEnvDeployer
	Market env.IDeployer `optional:"true"`
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	adminToken, err := depParams.VenusAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	venusEndpoint, err := depParams.Venus.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	gatewayEndpoint, err := depParams.Gateway.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	venusAuthEndpoint, err := depParams.VenusAuth.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	messagerEndpoint, err := depParams.Message.SvcEndpoint()
	if err != nil {
		return nil, err
	}
	marketEndpoint, err := depParams.Market.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	deployer, err := venussectormanager.DeployerFromConfig(depParams.K8sEnv, venussectormanager.Config{
		NodeUrl:     venusEndpoint.ToMultiAddr(),
		MessagerUrl: messagerEndpoint.ToMultiAddr(),
		MarketUrl:   marketEndpoint.ToMultiAddr(),
		GatewayUrl:  gatewayEndpoint.ToMultiAddr(),
		AuthUrl:     venusAuthEndpoint.ToHTTP(),
		AuthToken:   adminToken.MustString(),
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
