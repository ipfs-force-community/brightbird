package main

import (
	"context"
	"errors"

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

	Global     env.GlobalParams                  `ignore:"-" json:"global"`
	SophonAuth sophonauth.SophonAuthDeployReturn `json:"SophonAuth"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*venus.VenusDeployReturn, error) {
	var bootstrapPeers, ok = depParams.Global.CustomProperties["BootstrapPeer"].([]string)
	if !ok {
		return nil, errors.New("BootstrapPeer property not found")
	}
	return venus.DeployFromConfig(ctx, k8sEnv, venus.Config{
		BaseConfig: depParams.BaseConfig,
		VConfig: venus.VConfig{
			AuthUrl:        depParams.SophonAuth.SvcEndpoint.ToHTTP(),
			AdminToken:     depParams.SophonAuth.AdminToken,
			BootstrapPeers: bootstrapPeers,
			NetType:        depParams.NetType,
			Replicas:       depParams.Replicas,
		},
	})
}
