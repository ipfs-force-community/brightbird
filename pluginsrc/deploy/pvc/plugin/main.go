package main

import (
	"context"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/pvc"
)

func main() {
	plugin.SetupPluginFromStdin(pvc.PluginInfo, Exec)
}

type DepParams struct {
	env.BaseConfig
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*pvc.PvcReturn, error) {
	return pvc.DeployFromConfig(ctx, k8sEnv, pvc.Config{
		BaseConfig: depParams.BaseConfig,
	})
}
