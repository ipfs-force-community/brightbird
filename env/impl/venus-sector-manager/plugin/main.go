package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	venus_sector_manager "github.com/hunjixin/brightbird/env/impl/venus-sector-manager"
	"github.com/hunjixin/brightbird/types"
)

var Info = venus_sector_manager.PluginInfo

type DepParams struct {
	Params venus_sector_manager.Config `optional:"true"`

	Auth         env.IDeployer `svcname:"VenusAuth"`
	Venus        env.IDeployer `svcname:"Venus"`
	Message      env.IDeployer `svcname:"VenusMessager"`
	Gateway      env.IDeployer `svcname:"VenusGateway"`
	WalletDeploy env.IDeployer `svcname:VenusWallet`

	K8sEnv     *env.K8sEnvDeployer
	Market     env.IDeployer `optional:"true"`
	AdminToken types.AdminToken
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	deployer, err := venus_sector_manager.DeployerFromConfig(depParams.K8sEnv, venus_sector_manager.Config{
		NodeUrl:     depParams.Venus.SvcEndpoint().ToMultiAddr(),
		MessagerUrl: depParams.Message.SvcEndpoint().ToMultiAddr(),
		MarketUrl:   depParams.Market.SvcEndpoint().ToMultiAddr(),
		GatewayUrl:  depParams.Gateway.SvcEndpoint().ToMultiAddr(),
		AuthUrl:     depParams.Auth.SvcEndpoint().ToHttp(),
		AuthToken:   string(depParams.AdminToken),
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
