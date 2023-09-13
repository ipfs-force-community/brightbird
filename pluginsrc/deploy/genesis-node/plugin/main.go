package main

import (
	"context"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	genesisnode "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/genesis-node"
)

func main() {
	plugin.SetupPluginFromStdin(genesisnode.PluginInfo, Exec)
}

type DepParams struct {
	genesisnode.Config
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*genesisnode.GenesisReturn, error) {
	return genesisnode.DeployFromConfig(ctx, k8sEnv, genesisnode.Config{
		BaseConfig: depParams.BaseConfig,
	})
}
