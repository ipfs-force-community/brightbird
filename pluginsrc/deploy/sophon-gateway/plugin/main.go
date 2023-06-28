package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	sophongateway "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-gateway"
)

func main() {
	plugin.SetupPluginFromStdin(sophongateway.PluginInfo, Exec)
}

type DepParams struct {
	Params sophongateway.Config `optional:"true"`

	Auth env.IDeployer `svcname:"SophonAuth"`

	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	adminToken, err := depParams.Auth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	authEndpoint, err := depParams.Auth.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	deployer, err := sophongateway.DeployerFromConfig(depParams.K8sEnv, sophongateway.Config{
		AuthUrl:    authEndpoint.ToHTTP(),
		AdminToken: adminToken.MustString(),
	}, depParams.Params)
	if err != nil {
		return nil, err
	}

	err = deployer.Deploy(ctx)
	if err != nil {
		return nil, err
	}
	return deployer, nil
}
