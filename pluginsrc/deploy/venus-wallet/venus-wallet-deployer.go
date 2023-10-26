package venuswallet

import (
	"context"
	"embed"
	"fmt"

	"github.com/filecoin-project/venus-wallet/config"
	"github.com/ipfs-force-community/brightbird/env"
	venusutils "github.com/ipfs-force-community/brightbird/env/venus_utils"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	env.BaseConfig
	VConfig
}

type VConfig struct {
	GatewayUrl string `jsonschema:"-" json:"gatewayUrl"`
	UserToken  string `json:"userToken" jsonschema:"userToken" title:"UserToken" description:"token for connect with sophon gateway"`
}

type VenusWalletReturn struct { //nolint
	VConfig
	env.CommonDeployParams
}

type RenderParams struct {
	Config

	NameSpace string
	Registry  string
	Args      []string

	UniqueId string
}

var PluginInfo = types.PluginInfo{
	Name:       "venus-wallet",
	Version:    version.Version(),
	PluginType: types.Deploy,
	DeployPluginParams: types.DeployPluginParams{
		Repo:        "https://github.com/filecoin-project/venus-wallet.git",
		ImageTarget: "venus-wallet",
		BuildScript: `make docker-push TAG={{.Commit}} BUILD_DOCKER_PROXY={{.Proxy}} PRIVATE_REGISTRY={{.Registry}}`,
	},
	Description: "",
}

//go:embed venus-wallet
var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, cfg Config) (*VenusWalletReturn, error) {
	renderParams := RenderParams{
		NameSpace: k8sEnv.NameSpace(),
		Registry:  k8sEnv.Registry(),
		UniqueId:  env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), cfg.InstanceName),
		Config:    cfg,
	}
	//create configmap
	configMapCfg, err := f.Open("venus-wallet/venus-wallet-configmap.yaml")
	if err != nil {
		return nil, err
	}
	configMap, err := k8sEnv.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create deployment
	deployCfg, err := f.Open("venus-wallet/venus-wallet-statefulset.yaml")
	if err != nil {
		return nil, err
	}
	statefulSet, err := k8sEnv.RunStatefulSets(ctx, func(ctx context.Context, k8sEnv *env.K8sEnvDeployer) ([]corev1.Pod, error) {
		return GetPods(ctx, k8sEnv, cfg.InstanceName)
	}, deployCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create service
	svcCfg, err := f.Open("venus-wallet/venus-wallet-headless.yaml")
	if err != nil {
		return nil, err
	}
	svc, err := k8sEnv.RunService(ctx, svcCfg, renderParams)
	if err != nil {
		return nil, err
	}
	svcEndpoint, err := k8sEnv.WaitForServiceReady(ctx, svc, venusutils.VenusHealthCheck)
	if err != nil {
		return nil, err
	}
	return &VenusWalletReturn{
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

func Update(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params VenusWalletReturn, updateCfg config.Config) error {
	// cfgData, err := toml.Marshal(updateCfg)
	// if err != nil {
	// 	return err
	// }
	// err = k8sEnv.SetConfigMap(ctx, params.ConfigMapName, "config.toml", cfgData)
	// if err != nil {
	// 	return err
	// }

	// pods, err := GetPods(ctx, k8sEnv, params.InstanceName)
	// if err != nil {
	// 	return nil
	// }
	// for _, pod := range pods {
	// 	_, err = k8sEnv.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.venus_wallet/config.toml")
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return k8sEnv.UpdateStatefulSetsByName(ctx, params.StatefulSetName)
}

func GetPods(ctx context.Context, k8sEnv *env.K8sEnvDeployer, instanceName string) ([]corev1.Pod, error) {
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("venus-wallet-%s-pod", env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), instanceName)))
}
