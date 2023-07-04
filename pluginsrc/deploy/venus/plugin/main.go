package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	venus "github.com/hunjixin/brightbird/pluginsrc/deploy/venus"
)

func main() {
	plugin.SetupPluginFromStdin(venus.PluginInfo, Exec)
}

type DepParams struct {
	venus.Config

	Global     env.GlobalParams                  `json:"global"`
	SophonAuth sophonauth.SophonAuthDeployReturn `json:"SophonAuth"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*venus.VenusDeployReturn, error) {
	return venus.DeployFromConfig(ctx, k8sEnv, venus.Config{
		BaseConfig: depParams.BaseConfig,
		VConfig: venus.VConfig{
			AuthUrl:        depParams.SophonAuth.SvcEndpoint.ToHTTP(),
			AdminToken:     depParams.SophonAuth.AdminToken,
			BootstrapPeers: depParams.Global.BootrapPeers,
			NetType:        depParams.NetType,
			Replicas:       depParams.Replicas,
		},
	})
}
