package main

import (
	"context"
	"fmt"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	chainco "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-co"
)

func main() {
	plugin.SetupPluginFromStdin(chainco.PluginInfo, Exec)
}

type DepParams struct {
	Params chainco.Config `optional:"true"`
	K8sEnv *env.K8sEnvDeployer

	VenusDep   env.IDeployer `svcname:"Venus" description:"[Deploy]venus"`
	AuthDeploy env.IDeployer `svcname:"SophonAuth" description:"[Deploy]sophon-auth"`
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
	adminToken, err := depParams.AuthDeploy.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	podDNS := env.GetPodDNS(svc, pods...)
	podEndpoints := make([]string, len(podDNS))
	for index, dns := range podDNS {
		podEndpoints[index] = fmt.Sprintf("%s:/dns/%s/tcp/%d", adminToken, dns, svc.Spec.Ports[0].Port)
	}

	authEndpoint, err := depParams.AuthDeploy.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	deployer, err := chainco.DeployerFromConfig(depParams.K8sEnv, chainco.Config{
		Replicas:   1,
		AuthUrl:    authEndpoint.ToHTTP(),
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
