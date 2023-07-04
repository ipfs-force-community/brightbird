package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
)

func main() {
	plugin.SetupPluginFromStdin(sophonauth.PluginInfo, Exec)
}

type DepParams struct {
	sophonauth.Config `description:""`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*sophonauth.SophonAuthDeployReturn, error) {
	return sophonauth.DeployFromConfig(ctx, k8sEnv, depParams.Config)
}
