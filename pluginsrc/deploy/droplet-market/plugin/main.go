package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	dropletmarket "github.com/hunjixin/brightbird/pluginsrc/deploy/droplet-market"
)

var Info = dropletmarket.PluginInfo

func main() {
	plugin.SetupPluginFromStdin(dropletmarket.PluginInfo, Exec)
}

type DepParams struct {
	Params dropletmarket.Config `optional:"true"`

	Auth     env.IDeployer `svcname:"SophonAuth"`
	Venus    env.IDeployer `svcname:"Venus"`
	Messager env.IDeployer `svcname:"SophonMessager"`
	Gateway  env.IDeployer `svcname:"SophonGateway"`

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

	messagerEndpoint, err := depParams.Messager.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	gatewayEndpoint, err := depParams.Gateway.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	venusAuthEndpoint, err := depParams.Auth.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	deployer, err := dropletmarket.DeployerFromConfig(depParams.K8sEnv, dropletmarket.Config{
		NodeUrl:     venusEndpoint.ToMultiAddr(),
		GatewayUrl:  gatewayEndpoint.ToMultiAddr(),
		MessagerUrl: messagerEndpoint.ToMultiAddr(),
		AuthUrl:     venusAuthEndpoint.ToMultiAddr(),
		AuthToken:   adminToken.MustString(),
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
