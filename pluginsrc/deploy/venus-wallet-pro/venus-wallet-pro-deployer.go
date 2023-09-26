package venuswalletpro

import (
	"context"
	"embed"
	"fmt"

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
	MysqlDSN   string `jsonschema:"-"`

	Replicas  int    `json:"replicas"  jsonschema:"replicas" title:"replicas" default:"1" require:"true" description:"number of replicas"`
	UserToken string `json:"userToken" jsonschema:"userToken" title:"UserToken" description:"token for connect with sophon gateway"`
}

type VenusWalletProDeployReturn struct { //nolint
	VConfig
	env.CommonDeployParams
}

type RenderParams struct {
	Config

	Registry string
	Args     []string

	UniqueId string
}

var PluginInfo = types.PluginInfo{
	Name:       "venus-wallet-pro",
	Version:    version.Version(),
	PluginType: types.Deploy,
	DeployPluginParams: types.DeployPluginParams{
		Repo:        "",
		ImageTarget: "venus-wallet-pro",
		BuildScript: `make docker-push TAG={{.Commit}} BUILD_DOCKER_PROXY={{.Proxy}} PRIVATE_REGISTRY={{.Registry}}`,
	},
	Description: "",
}

var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, cfg Config) (*VenusWalletProDeployReturn, error) {
	cfg.MysqlDSN = k8sEnv.FormatMysqlConnection("venus-wallet-pro-" + env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), cfg.InstanceName))
	renderParams := RenderParams{
		Registry: k8sEnv.Registry(),
		UniqueId: env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), cfg.InstanceName),
		Config:   cfg,
	}
	//create database
	err := k8sEnv.ResourceMgr().EnsureDatabase(cfg.MysqlDSN)
	if err != nil {
		return nil, err
	}

	//create configmap
	configMapCfg, err := f.Open("venus-wallet-pro/venus-wallet-pro-configmap.yaml")
	if err != nil {
		return nil, err
	}
	configMap, err := k8sEnv.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create deployment
	deployCfg, err := f.Open("venus-wallet-pro/venus-wallet-pro-statefulset.yaml")
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
	svcCfg, err := f.Open("venus-wallet-pro/venus-wallet-pro-headless.yaml")
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
	return &VenusWalletProDeployReturn{
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
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("venus-wallet-pro-%s-pod", env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), instanceName)))
}
