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
	BootstrapPeers []string                          `json:"bootstrapPeers" jsonschema:"bootstrapPeers" title:"BootstrapPeers" require:"true" description:"config boot peers"`
	Auth           sophonauth.SophonAuthDeployReturn `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`

	GenesisStorage  string `json:"genesisStorage"  jsonschema:"genesisStorage" title:"GenesisStorage" default:"" require:"true" description:"used genesis file"`
	SnapshotStorage string `json:"snapshotStorage"  jsonschema:"snapshotStorage" title:"SnapshotStorage" default:"" require:"true" description:"used to read snapshot file"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*venus.VenusDeployReturn, error) {
	return venus.DeployFromConfig(ctx, k8sEnv, venus.Config{
		BaseConfig: depParams.BaseConfig,
		VConfig: venus.VConfig{
			GenesisStorage:  depParams.GenesisStorage,
			SnapshotStorage: depParams.SnapshotStorage,
			AuthUrl:         depParams.Auth.SvcEndpoint.ToHTTP(),
			AdminToken:      depParams.Auth.AdminToken,
			BootstrapPeers:  depParams.BootstrapPeers,
			NetType:         depParams.NetType,
			Replicas:        depParams.Replicas,
		},
	})
}
