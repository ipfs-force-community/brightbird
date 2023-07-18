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

	Auth     sophonauth.SophonAuthDeployReturn   `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Venus    venus.VenusDeployReturn             `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	Gateway  sophongateway.SophonGatewayReturn   `json:"SophonGateway"  jsonschema:"SophonGateway"  title:"SophonGateway" require:"true" description:"gateway deploy return"`
	Messager sophonmessager.SophonMessagerReturn `json:"SophonMessager"  jsonschema:"SophonMessager"  title:"Sophon Messager" require:"true" description:"messager return"`

	DamoclesManager damoclesmanager.DamoclesManagerReturn `json:"DamoclesManager" jsonschema:"DamoclesManager" title:"Damocles Manager" description:"damocles manager return" require:"true"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*damoclesworker.DropletMarketDeployReturn, error) {
	return damoclesworker.DeployFromConfig(ctx, k8sEnv, damoclesworker.Config{
		BaseConfig: depParams.BaseConfig,
		VConfig: damoclesworker.VConfig{
			DamoclesManagerURL: depParams.DamoclesManager.SvcEndpoint.ToMultiAddr(),
			MarketToken:        depParams.MarketToken,
			MinerAddress:       depParams.MinerAddress,
		},
	})
}
