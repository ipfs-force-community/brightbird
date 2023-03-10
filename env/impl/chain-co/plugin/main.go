package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hunjixin/brightbird/env"
	chain_co "github.com/hunjixin/brightbird/env/impl/chain-co"
	"github.com/hunjixin/brightbird/types"
)

var Info = chain_co.PluginInfo

type DepParams struct {
	Params          json.RawMessage `optional:"true"` //todo use params direct
	K8sEnv          *env.K8sEnvDeployer
	VenusDep        env.IVenusDeployer
	VenusAuthDeploy env.IVenusAuthDeployer
	AdminToken      types.AdminToken
}

func Exec(ctx context.Context, depParams DepParams) (env.IChainCoDeployer, error) {
	podDNS := env.GetPodDNS(depParams.VenusDep.Svc(), depParams.VenusDep.Pods()...)
	podEndpoints := make([]string, len(podDNS))
	for index, dns := range podDNS {
		podEndpoints[index] = fmt.Sprintf("%s:/dns/%s/tcp/%d", depParams.AdminToken, dns, depParams.VenusDep.Svc().Spec.Ports[0].Port)
	}

	deployer, err := chain_co.DeployerFromConfig(depParams.K8sEnv, chain_co.Config{
		Replicas: 1,
		AuthUrl:  depParams.VenusAuthDeploy.SvcEndpoint().ToHttp(),
		Nodes:    podEndpoints,
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
