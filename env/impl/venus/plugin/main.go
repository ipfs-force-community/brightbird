package main

import (
	"context"
	"encoding/json"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/impl/venus"
	"github.com/hunjixin/brightbird/types"
)

var Info = venus.PluginInfo

type DepParams struct {
	Params          json.RawMessage `optional:"true"`
	K8sEnv          *env.K8sEnvDeployer
	VenusAuthDeploy env.IVenusAuthDeployer

	AdminToken     types.AdminToken
	BootstrapPeers types.BootstrapPeers
}

func Exec(ctx context.Context, depParams DepParams) (env.IVenusDeployer, error) {
	deployer, err := venus.DeployerFromConfig(depParams.K8sEnv, venus.Config{
		AuthUrl:        depParams.VenusAuthDeploy.SvcEndpoint().ToHttp(),
		AdminToken:     string(depParams.AdminToken),
		BootstrapPeers: depParams.BootstrapPeers,
	}, depParams.Params)
	if err != nil {
		return nil, err
	}

	err = deployer.Deploy(ctx)
	if err != nil {
		return nil, err
	}

	err = env.SyncWait(ctx, depParams.K8sEnv, deployer, string(depParams.AdminToken))
	if err != nil {
		return nil, err
	}
	return deployer, nil
}
