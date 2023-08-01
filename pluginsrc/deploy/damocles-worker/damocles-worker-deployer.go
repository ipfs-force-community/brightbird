package damoclesworker

import (
	"context"
	"embed"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
	corev1 "k8s.io/api/core/v1"
)

type DamoclesWorkerConfig string //nolint

type Config struct {
	env.BaseConfig
	VConfig
}

type VConfig struct {
	DamoclesManagerURL string `jsonschema:"-" json:"damoclesManagerURL" title:"DamoclesManagerURL"  require:"true" `

	MarketToken  string          `json:"marketToken" jsonschema:"marketToken" title:"Market Token"  require:"true" `
	MinerAddress address.Address `json:"minerAddress" jsonschema:"minerAddress" title:"Miner Address"  require:"true" `
}

type DropletMarketDeployReturn struct {
	VConfig
	env.CommonDeployParams
}

type RenderParams struct {
	Config

	NameSpace       string
	PrivateRegistry string
	Args            []string

	TestID string
}

var PluginInfo = types.PluginInfo{
	Name:        "damocles-worker",
	Version:     version.Version(),
	PluginType:  types.Deploy,
	Repo:        "https://github.com/ipfs-force-community/damocles.git",
	ImageTarget: "damocles-worker",
	Description: "",
}

var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, cfg Config) (*DropletMarketDeployReturn, error) {
	renderParams := RenderParams{
		NameSpace:       k8sEnv.NameSpace(),
		PrivateRegistry: k8sEnv.PrivateRegistry(),
		TestID:          k8sEnv.TestID(),
		Config:          cfg,
	}

	// create configMap
	configMapCfg, err := f.Open("damocles-worker/damocles-worker-configmap.yaml")
	if err != nil {
		return nil, err
	}
	configMap, err := k8sEnv.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return nil, err
	}

	// create deployment
	deployCfg, err := f.Open("damocles-worker/damocles-worker-statefulset.yaml")
	if err != nil {
		return nil, err
	}
	statefulSet, err := k8sEnv.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return nil, err
	}

	// create service
	svcCfg, err := f.Open("damocles-worker/damocles-worker-headless.yaml")
	if err != nil {
		return nil, err
	}

	svc, err := k8sEnv.RunService(ctx, svcCfg, renderParams)
	if err != nil {
		return nil, err
	}

	svcEndpoint, err := k8sEnv.WaitForServiceReady(ctx, svc)
	if err != nil {
		return nil, err
	}

	return &DropletMarketDeployReturn{
		VConfig: cfg.VConfig,
		CommonDeployParams: env.CommonDeployParams{
			BaseConfig:      cfg.BaseConfig,
			DeployName:      PluginInfo.Name,
			StatefulSetName: statefulSet.GetName(),
			ConfigMapName:   configMap.GetName(),
			SVCName:         svc.GetName(),
			SvcEndpoint:     svcEndpoint,
		},
	}, nil
}

func GetPods(ctx context.Context, k8sEnv *env.K8sEnvDeployer, instanceName string) ([]corev1.Pod, error) {
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("damocles-worker-%s-pod", env.UniqueId(k8sEnv.TestID(), instanceName)))
}
