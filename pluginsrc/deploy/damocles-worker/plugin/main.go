package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	damoclesmanager "github.com/hunjixin/brightbird/pluginsrc/deploy/damocles-manager"
	damoclesworker "github.com/hunjixin/brightbird/pluginsrc/deploy/damocles-worker"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	sophongateway "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-gateway"
	sophonmessager "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-messager"
	"github.com/hunjixin/brightbird/pluginsrc/deploy/venus"
)

func main() {
	plugin.SetupPluginFromStdin(damoclesworker.PluginInfo, Exec)
}

type DepParams struct {
	damoclesworker.Config

	Auth     sophonauth.SophonAuthDeployReturn   `json:"SophonAuth" description:"sophon auth return"`
	Venus    venus.VenusDeployReturn             `json:"Venus" description:"venus return"`
	Gateway  sophongateway.SophonGatewayReturn   `json:"SophonGateway" description:"gateway return"`
	Messager sophonmessager.SophonMessagerReturn `json:"SophonMessager" description:"messager return"`

	SectorManager damoclesmanager.DamoclesManagerReturn `json:"DamoclesManager" description:"damocles manager return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*damoclesworker.DropletMarketDeployReturn, error) {
	return damoclesworker.DeployFromConfig(ctx, k8sEnv, damoclesworker.Config{
		BaseConfig: depParams.BaseConfig,
		VConfig: damoclesworker.VConfig{
			DamoclesManagerURL: depParams.SectorManager.SvcEndpoint.ToMultiAddr(),
			MarketToken:        depParams.MarketToken,
			MinerAddress:       depParams.MinerAddress,
		},
	})
}
