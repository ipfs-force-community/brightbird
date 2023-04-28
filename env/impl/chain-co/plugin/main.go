package main

import (
	"context"
	"fmt"

	"github.com/hunjixin/brightbird/env"
	chain_co "github.com/hunjixin/brightbird/env/impl/chain-co"
	"github.com/hunjixin/brightbird/types"
)

var Info = chain_co.PluginInfo

type DepParams struct {
	Params chain_co.Config `optional:"true"`
	K8sEnv *env.K8sEnvDeployer

	VenusDep        env.IVenusDeployer
	VenusAuthDeploy env.IVenusAuthDeployer
	AdminToken      types.AdminToken
}

func Exec(ctx context.Context, depParams DepParams) (env.IChainCoDeployer, error) {
	pods, err := depParams.VenusDep.Pods(ctx)
	if err != nil {
		return nil, err
	}
	svc, err := depParams.VenusDep.Svc(ctx)
	if err != nil {
		return nil, err
	}
	podDNS := env.GetPodDNS(svc, pods...)
	podEndpoints := make([]string, len(podDNS))
	for index, dns := range podDNS {
		podEndpoints[index] = fmt.Sprintf("%s:/dns/%s/tcp/%d", depParams.AdminToken, dns, svc.Spec.Ports[0].Port)
	}

	deployer, err := chain_co.DeployerFromConfig(depParams.K8sEnv, chain_co.Config{
		Replicas:   1,
		AuthUrl:    depParams.VenusAuthDeploy.SvcEndpoint().ToHttp(),
		AdminToken: depParams.Params.AdminToken,
		Nodes:      podEndpoints,
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
