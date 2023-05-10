package main

import (
	"context"
	"errors"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/types"
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

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	return &VenusTestDeploy{}, nil
}

type VenusTestDeploy struct {
}

func (dep *VenusTestDeploy) Name() string {
	return "venus-test"
}
func (dep *VenusTestDeploy) Pods(ctx context.Context) ([]corev1.Pod, error) {
	return nil, nil
}

func (dep *VenusTestDeploy) Svc(_ context.Context) (*corev1.Service, error) {
	return nil, nil
}

func (dep *VenusTestDeploy) StatefulSet(ctx context.Context) (*appv1.StatefulSet, error) {
	return nil, nil
}
func (dep *VenusTestDeploy) SvcEndpoint() types.Endpoint {
	return ""
}

func (deployer *VenusTestDeploy) Param(key string) (interface{}, error) {
	return nil, errors.New("no params")
}

func (dep *VenusTestDeploy) Deploy(ctx context.Context) (err error) {
	return nil
}

func (dep *VenusTestDeploy) GetConfig(ctx context.Context) (interface{}, error) {
	return nil, nil
}
func (dep *VenusTestDeploy) Update(ctx context.Context, updateCfg interface{}) error {
	return nil
}
