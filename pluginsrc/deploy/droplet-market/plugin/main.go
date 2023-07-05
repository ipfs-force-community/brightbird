package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	dropletmarket "github.com/hunjixin/brightbird/pluginsrc/deploy/droplet-market"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	sophongateway "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-gateway"
	sophonmessager "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-messager"
	"github.com/hunjixin/brightbird/pluginsrc/deploy/venus"
)

var Info = dropletmarket.PluginInfo

func main() {
	plugin.SetupPluginFromStdin(dropletmarket.PluginInfo, Exec)
}

type DepParams struct {
	dropletmarket.Config

	Auth     sophonauth.SophonAuthDeployReturn   `json:"SophonAuth" description:"sophon auth return"`
	Venus    venus.VenusDeployReturn             `json:"Venus" description:"venus return"`
	Gateway  sophongateway.SophonGatewayReturn   `json:"SophonGateway" description:"gateway return"`
	Messager sophonmessager.SophonMessagerReturn `json:"SophonMessager" description:"messager return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*dropletmarket.DropletMarketDeployReturn, error) {
	return dropletmarket.DeployFromConfig(ctx, k8sEnv, dropletmarket.Config{
		BaseConfig: depParams.BaseConfig,
		VConfig: dropletmarket.VConfig{
			NodeUrl:     depParams.Venus.SvcEndpoint.ToMultiAddr(),
			GatewayUrl:  depParams.Gateway.SvcEndpoint.ToMultiAddr(),
			MessagerUrl: depParams.Messager.SvcEndpoint.ToMultiAddr(),
			AuthUrl:     depParams.Auth.SvcEndpoint.ToHTTP(),

			UserToken: depParams.UserToken,
			UseMysql:  depParams.UseMysql,
		},
	})
}
