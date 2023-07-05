package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	damoclesmanager "github.com/hunjixin/brightbird/pluginsrc/deploy/damocles-manager"
	dropletmarket "github.com/hunjixin/brightbird/pluginsrc/deploy/droplet-market"

	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	sophongateway "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-gateway"
	sophonmessager "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-messager"
	"github.com/hunjixin/brightbird/pluginsrc/deploy/venus"
)

func main() {
	plugin.SetupPluginFromStdin(damoclesmanager.PluginInfo, Exec)
}

type DepParams struct {
	damoclesmanager.Config

	Auth     sophonauth.SophonAuthDeployReturn   `json:"SophonAuth" description:"sophon auth return"`
	Venus    venus.VenusDeployReturn             `json:"Venus" description:"venus return"`
	Gateway  sophongateway.SophonGatewayReturn   `json:"SophonGateway" description:"gateway return"`
	Messager sophonmessager.SophonMessagerReturn `json:"SophonMessager" description:"messager return"`

	Market dropletmarket.DropletMarketDeployReturn `json:"DropletMarket" description:"droplet market return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*damoclesmanager.DamoclesManagerReturn, error) {
	return damoclesmanager.DeployFromConfig(ctx, k8sEnv, damoclesmanager.Config{
		BaseConfig: depParams.BaseConfig,
		VConfig: damoclesmanager.VConfig{
			NodeUrl:             depParams.Auth.SvcEndpoint.ToMultiAddr(),
			MessagerUrl:         depParams.Messager.SvcEndpoint.ToMultiAddr(),
			MarketUrl:           depParams.Market.SvcEndpoint.ToMultiAddr(),
			GatewayUrl:          depParams.Gateway.SvcEndpoint.ToMultiAddr(),
			AuthUrl:             depParams.Auth.SvcEndpoint.ToHTTP(),
			UserToken:           depParams.UserToken,
			MinerAddress:        depParams.MinerAddress,
			SenderWalletAddress: depParams.SenderWalletAddress,
		},
	})
}
