package main

import (
	"context"
	"fmt"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	chainco "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-co"
	"github.com/hunjixin/brightbird/pluginsrc/deploy/venus"
	corev1 "k8s.io/api/core/v1"
)

func main() {
	plugin.SetupPluginFromStdin(chainco.PluginInfo, Exec)
}

type DepParams struct {
	chainco.Config

	Daemon     env.CommonDeployParams            `json:"Daemon" description:"[Deploy]venus/lotus/sophonco daemon"`
	SophonAuth sophonauth.SophonAuthDeployReturn `json:"SophonAuth"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*chainco.SophonCoDeployReturn, error) {

	var pods []corev1.Pod
	var err error
	switch depParams.Daemon.DeployName {
	case venus.PluginInfo.Name:
		pods, err = venus.GetPods(ctx, k8sEnv, depParams.Daemon.InstanceName)
		if err != nil {
			return nil, err
		}
	}

	svc, err := k8sEnv.GetSvc(ctx, depParams.Daemon.SVCName)
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
			Replicas:   1,
			AuthUrl:    depParams.SophonAuth.SvcEndpoint.ToHTTP(),
			AdminToken: depParams.SophonAuth.AdminToken,
			Nodes:      podEndpoints,
		},
	})
}
