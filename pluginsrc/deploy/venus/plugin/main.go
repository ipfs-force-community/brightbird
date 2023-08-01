package main

import (
	"context"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	venus "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
)

func main() {
	plugin.SetupPluginFromStdin(venus.PluginInfo, Exec)
}

type DepParams struct {
	venus.Config

	Global env.GlobalParams                  `jsonschema:"-" json:"global"`
	Auth   sophonauth.SophonAuthDeployReturn `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*venus.VenusDeployReturn, error) {
	var bootstrapPeers []string
	err := depParams.Global.GetProperty("BootstrapPeer", &bootstrapPeers)
	if err != nil {
		return nil, err
	}
	return venus.DeployFromConfig(ctx, k8sEnv, venus.Config{
		BaseConfig: depParams.BaseConfig,
		VConfig: venus.VConfig{
			AuthUrl:        depParams.Auth.SvcEndpoint.ToHTTP(),
			AdminToken:     depParams.Auth.AdminToken,
			BootstrapPeers: bootstrapPeers,
			NetType:        depParams.NetType,
			Replicas:       depParams.Replicas,
		},
	})
}
