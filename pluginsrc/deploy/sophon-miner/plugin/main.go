package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	sophongateway "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-gateway"
	sophonminer "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-miner"
	"github.com/hunjixin/brightbird/pluginsrc/deploy/venus"
)

func main() {
	plugin.SetupPluginFromStdin(sophonminer.PluginInfo, Exec)
}

type DepParams struct {
	sophonminer.Config

	Auth    sophonauth.SophonAuthDeployReturn `json:"SophonAuth" description:"sophon auth return"`
	Venus   venus.VenusDeployReturn           `json:"Venus" description:"venus return"`
	Gateway sophongateway.SophonGatewayReturn `json:"SophonGateway" description:"gateway return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*sophonminer.SophonMinerDeployReturn, error) {
	return sophonminer.DeployFromConfig(ctx, k8sEnv, sophonminer.Config{
		BaseConfig: depParams.BaseConfig,
		VConfig: sophonminer.VConfig{
			NodeUrl:    depParams.Venus.SvcEndpoint.ToMultiAddr(),
			GatewayUrl: depParams.Gateway.SvcEndpoint.ToMultiAddr(),
			AuthUrl:    depParams.Auth.SvcEndpoint.ToHTTP(),
			AuthToken:  depParams.Auth.AdminToken,
			UseMysql:   depParams.UseMysql,
		},
	})
}
