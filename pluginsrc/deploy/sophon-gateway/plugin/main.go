package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	sophongateway "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-gateway"
)

func main() {
	plugin.SetupPluginFromStdin(sophongateway.PluginInfo, Exec)
}

type DepParams struct {
	sophongateway.Config

	Auth sophonauth.SophonAuthDeployReturn `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*sophongateway.SophonGatewayReturn, error) {
	return sophongateway.DeployFromConfig(ctx, k8sEnv, sophongateway.Config{
		BaseConfig: depParams.BaseConfig,
		VConfig: sophongateway.VConfig{
			AuthUrl:    depParams.Auth.SvcEndpoint.ToHTTP(),
			AdminToken: depParams.Auth.AdminToken,
			Replicas:   depParams.Replicas,
		},
	})
}
