package main

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	damoclesmanager "github.com/hunjixin/brightbird/pluginsrc/deploy/damocles-manager"
)

func main() {
	plugin.SetupPluginFromStdin(damoclesmanager.PluginInfo, Exec)
}

type DepParams struct {
	Params damoclesmanager.Config `optional:"true"`

	Auth         env.IDeployer `svcname:"SophonAuth"`
	Venus        env.IDeployer `svcname:"Venus"`
	Message      env.IDeployer `svcname:"SophonMessager"`
	Gateway      env.IDeployer `svcname:"SophonGateway"`
	WalletDeploy env.IDeployer `svcname:"VenusWallet"`

	K8sEnv *env.K8sEnvDeployer
	Market env.IDeployer `optional:"true"`

	Miner env.IExec `svcname:"MinerInfo"`
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	adminToken, err := depParams.Auth.Param("AdminToken")
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

	authEndpoint, err := depParams.Auth.SvcEndpoint()
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

	minerP, err := depParams.Miner.Param("Miner")
	if err != nil {
		return nil, err
	}
	minerAddr, err := env.UnmarshalJSON[address.Address](minerP.Raw())
	if err != nil {
		return nil, err
	}

	workerP, err := depParams.Miner.Param("Worker")
	if err != nil {
		return nil, err
	}
	workerAddr, err := env.UnmarshalJSON[address.Address](workerP.Raw())
	if err != nil {
		return nil, err
	}

	deployer, err := damoclesmanager.DeployerFromConfig(depParams.K8sEnv, damoclesmanager.Config{
		NodeUrl:             venusEndpoint.ToMultiAddr(),
		MessagerUrl:         messagerEndpoint.ToMultiAddr(),
		MarketUrl:           marketEndpoint.ToMultiAddr(),
		GatewayUrl:          gatewayEndpoint.ToMultiAddr(),
		AuthUrl:             authEndpoint.ToHTTP(),
		AuthToken:           adminToken.MustString(),
		MinerAddress:        minerAddr,
		SenderWalletAddress: workerAddr,
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
