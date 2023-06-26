package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	sophonmessager "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-messager"
)

func main() {
	plugin.SetupPluginFromStdin(sophonmessager.PluginInfo, Exec)
}

type DepParams struct {
	Params sophonmessager.Config `optional:"true"`

	Auth    env.IDeployer `svcname:"SophonAuth"`
	Venus   env.IDeployer `svcname:"Venus"`
	Gateway env.IDeployer `svcname:"SophonGateway"`

	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	adminToken, err := depParams.Auth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	venusEndpoint, err := depParams.Venus.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	gatewayEndpoint, err := depParams.Gateway.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	authEndpoint, err := depParams.Auth.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	deployer, err := sophonmessager.DeployerFromConfig(depParams.K8sEnv, sophonmessager.Config{
		NodeUrl:    venusEndpoint.ToMultiAddr(),
		GatewayUrl: gatewayEndpoint.ToMultiAddr(),
		AuthUrl:    authEndpoint.ToHTTP(),
		AuthToken:  adminToken.MustString(),
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
