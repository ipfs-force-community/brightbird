package main

import (
	"context"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	sophongateway "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-gateway"
	sophonmessager "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-messager"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
)

func main() {
	plugin.SetupPluginFromStdin(sophonmessager.PluginInfo, Exec)
}

type DepParams struct {
	sophonmessager.Config

	Auth    sophonauth.SophonAuthDeployReturn `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Venus   venus.VenusDeployReturn           `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	Gateway sophongateway.SophonGatewayReturn `json:"SophonGateway"  jsonschema:"SophonGateway"  title:"SophonGateway" require:"true" description:"gateway deploy return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*sophonmessager.SophonMessagerReturn, error) {
	return sophonmessager.DeployFromConfig(ctx, k8sEnv, sophonmessager.Config{
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
