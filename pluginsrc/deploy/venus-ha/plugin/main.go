package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	venus "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-ha"
	"github.com/hunjixin/brightbird/types"
)

func main() {
	plugin.SetupPluginFromStdin(venus.PluginInfo, Exec)
}

type DepParams struct {
	Params venus.Config `optional:"true"`

	VenusAuth env.IDeployer `svcname:"VenusAuth"`

	K8sEnv         *env.K8sEnvDeployer
	BootstrapPeers types.BootstrapPeers
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

	deployer, err := venus.DeployerFromConfig(depParams.K8sEnv, venus.Config{
		AuthUrl:        authEndpoint.ToHTTP(),
		AdminToken:     adminToken.MustString(),
		BootstrapPeers: depParams.BootstrapPeers,
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
