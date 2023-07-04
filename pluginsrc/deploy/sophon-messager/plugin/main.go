package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	sophongateway "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-gateway"
	sophonmessager "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-messager"
	"github.com/hunjixin/brightbird/pluginsrc/deploy/venus"
)

func main() {
	plugin.SetupPluginFromStdin(sophonmessager.PluginInfo, Exec)
}

type DepParams struct {
	sophonmessager.Config

	Auth    sophonauth.SophonAuthDeployReturn `json:"SophonAuth" description:"sophon auth return"`
	Venus   venus.VenusDeployReturn           `json:"Venus" description:"venus return"`
	Gateway sophongateway.SophonGatewayReturn `json:"SophonGateway" description:"gateway return"`

	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (*sophonmessager.SophonMessagerReturn, error) {
	return sophonmessager.DeployFromConfig(ctx, depParams.K8sEnv, sophonmessager.Config{
		BaseConfig: depParams.BaseConfig,
		VConfig: sophonmessager.VConfig{
			NodeUrl:    depParams.Venus.SvcEndpoint.ToMultiAddr(),
			GatewayUrl: depParams.Gateway.SvcEndpoint.ToMultiAddr(),
			AuthUrl:    depParams.Auth.SvcEndpoint.ToHTTP(),
			AuthToken:  depParams.Auth.AdminToken,
			Replicas:   depParams.Replicas,
		},
	})
}
