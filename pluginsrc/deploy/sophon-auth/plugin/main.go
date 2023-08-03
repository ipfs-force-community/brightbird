package main

import (
	"context"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
)

func main() {
	plugin.SetupPluginFromStdin(sophonauth.PluginInfo, Exec)
}

type DepParams struct {
	sophonauth.Config
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*sophonauth.SophonAuthDeployReturn, error) {
	return sophonauth.DeployFromConfig(ctx, k8sEnv, depParams.Config)
}
