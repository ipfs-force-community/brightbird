package main

import (
	"context"
	"fmt"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	chainco "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-co"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
	corev1 "k8s.io/api/core/v1"
)

func main() {
	plugin.SetupPluginFromStdin(chainco.PluginInfo, Exec)
}

type DepParams struct {
	chainco.Config

	Venus env.CommonDeployParams            `json:"Venus"  jsonschema:"Venus Daemon" title:"Venus Daemon" require:"true" description:"[Deploy]venus/lotus/sophonco daemon"`
	Auth  sophonauth.SophonAuthDeployReturn `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*chainco.SophonCoDeployReturn, error) {

	var pods []corev1.Pod
	var err error
	switch depParams.Venus.DeployName {
	case venus.PluginInfo.Name:
		pods, err = venus.GetPods(ctx, k8sEnv, depParams.Venus.InstanceName)
		if err != nil {
			return nil, err
		}
	}

	svc, err := k8sEnv.GetSvc(ctx, depParams.Venus.SVCName)
	if err != nil {
		return nil, err
	}

	podDNS := env.GetPodDNS(svc, pods...)
	podEndpoints := make([]string, len(podDNS))
	for index, dns := range podDNS {

		podEndpoints[index] = fmt.Sprintf("%s:/dns/%s/tcp/%d", depParams.AdminToken, dns, svc.Spec.Ports[0].Port)
	}

	return chainco.DeployFromConfig(ctx, k8sEnv, chainco.Config{
		BaseConfig: depParams.BaseConfig,
		VConfig: chainco.VConfig{
			AuthUrl:    depParams.Auth.SvcEndpoint.ToHTTP(),
			AdminToken: depParams.Auth.AdminToken,
			Nodes:      podEndpoints,
			Replicas:   depParams.Replicas,
		},
	})
}
