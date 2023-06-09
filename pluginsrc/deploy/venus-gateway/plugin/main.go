package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	venusgateway "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-gateway"
)

func main() {
	plugin.SetupPluginFromStdin(venusgateway.PluginInfo, Exec)
}

type DepParams struct {
	Params venusgateway.Config `optional:"true"`

	VenusAuth env.IDeployer `svcname:"VenusAuth"`

	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	adminToken, err := depParams.VenusAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	authEndpoint, err := depParams.VenusAuth.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	deployer, err := venusgateway.DeployerFromConfig(depParams.K8sEnv, venusgateway.Config{
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
