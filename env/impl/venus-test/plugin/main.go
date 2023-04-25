package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var Info = types.PluginInfo{
	Name:        "venus-test",
	Version:     version.Version(),
	Category:    types.Deploy,
	Repo:        "https://github.com/ipfs-force-community/webhooktest.git",
	ImageTarget: "venus-webhook-test",
	Description: "ignore, just for test webhook",
}

type Config struct {
	env.BaseConfig
}

type DepParams struct {
	Params Config `optional:"true"`
	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.ITestDeployer, error) {
	return &VenusTestDeploy{}, nil
}

type VenusTestDeploy struct {
}

func (dep *VenusTestDeploy) Name() string {
	return "venus-test"
}
func (dep *VenusTestDeploy) Pods() []corev1.Pod {
	return nil
}
func (dep *VenusTestDeploy) Deployment() []*appv1.Deployment {
	return nil
}
func (dep *VenusTestDeploy) Svc() *corev1.Service {
	return nil
}
func (dep *VenusTestDeploy) SvcEndpoint() types.Endpoint {
	return ""
}
func (dep *VenusTestDeploy) Deploy(ctx context.Context) (err error) {
	return nil
}
