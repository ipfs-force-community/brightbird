package main

import (
	"context"
	"fmt"

	"github.com/hunjixin/brightbird/env"
	venus_gateway "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-gateway"
)

var Info = venus_gateway.PluginInfo

type DepParams struct {
	Params venus_gateway.Config `optional:"true"`

	VenusAuth env.IDeployer `svcname:"VenusAuth"`

	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	adminToken, err := depParams.VenusAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}
	fmt.Println("ttttt ggg", adminToken)
	deployer, err := venus_gateway.DeployerFromConfig(depParams.K8sEnv, venus_gateway.Config{
		AuthUrl:    depParams.VenusAuth.SvcEndpoint().ToHttp(),
		AdminToken: adminToken.(string),
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
