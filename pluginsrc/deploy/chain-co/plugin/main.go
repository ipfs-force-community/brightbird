package main

import (
	"context"
	"fmt"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	chainco "github.com/hunjixin/brightbird/pluginsrc/deploy/chain-co"
)

func main() {
	plugin.SetupPluginFromStdin(chainco.PluginInfo, Exec)
}

type DepParams struct {
	Params chainco.Config `optional:"true"`
	K8sEnv *env.K8sEnvDeployer

	VenusDep        env.IDeployer `svcname:"Venus" description:"[Deploy]venus"`
	VenusAuthDeploy env.IDeployer `svcname:"VenusAuth" description:"[Deploy]venus-auth"`
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	pods, err := depParams.VenusDep.Pods(ctx)
	if err != nil {
		return nil, err
	}
	svc, err := depParams.VenusDep.Svc(ctx)
	if err != nil {
		return nil, err
	}
	adminToken, err := depParams.VenusAuthDeploy.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	podDNS := env.GetPodDNS(svc, pods...)
	podEndpoints := make([]string, len(podDNS))
	for index, dns := range podDNS {
		podEndpoints[index] = fmt.Sprintf("%s:/dns/%s/tcp/%d", adminToken, dns, svc.Spec.Ports[0].Port)
	}

	venusAuthEndpoint, err := depParams.VenusAuthDeploy.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	deployer, err := chainco.DeployerFromConfig(depParams.K8sEnv, chainco.Config{
		Replicas:   1,
		AuthUrl:    venusAuthEndpoint.ToHTTP(),
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
