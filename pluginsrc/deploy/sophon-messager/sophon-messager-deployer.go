package sophonmessager

import (
	"context"
	"embed"
	"fmt"

	"github.com/ipfs-force-community/brightbird/env"
	venusutils "github.com/ipfs-force-community/brightbird/env/venus_utils"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
	"github.com/ipfs-force-community/sophon-messager/config"
	"github.com/pelletier/go-toml"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	env.BaseConfig
	VConfig
}

type VConfig struct {
	NodeUrl    string `jsonschema:"-" json:"nodeUrl"`
	GatewayUrl string `jsonschema:"-" json:"gatewayUrl"`
	AuthUrl    string `jsonschema:"-" json:"authUrl"`
	AuthToken  string `jsonschema:"-" json:"authToken"`
	MysqlDSN   string `jsonschema:"-" json:"mysqlDSN"`

	Replicas int `json:"replicas"  jsonschema:"replicas" title:"replicas" default:"1" require:"true" description:"number of replicas"`
}

type SophonMessagerReturn struct { //nolint
	VConfig
	env.CommonDeployParams
}

type RenderParams struct {
	Config

	NameSpace string
	Registry  string
	Args      []string
	UniqueId  string
}

var PluginInfo = types.PluginInfo{
	Name:       "sophon-messager",
	Version:    version.Version(),
	PluginType: types.Deploy,
	DeployPluginParams: types.DeployPluginParams{
		Repo:        "https://github.com/ipfs-force-community/sophon-messager.git",
		ImageTarget: "sophon-messager",
		BuildScript: `make docker-push TAG={{.Commit}} BUILD_DOCKER_PROXY={{.Proxy}} PRIVATE_REGISTRY={{.Registry}}`,
	},
	Description: "",
}

//go:embed sophon-messager
var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, cfg Config) (*SophonMessagerReturn, error) {
	cfg.MysqlDSN = k8sEnv.FormatMysqlConnection("sophon-messager-" + env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), cfg.InstanceName))
	renderParams := RenderParams{
		NameSpace: k8sEnv.NameSpace(),
		Registry:  k8sEnv.Registry(),
		UniqueId:  env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), cfg.InstanceName),
		Config:    cfg,
	}

	//create database
	err := k8sEnv.ResourceMgr().EnsureDatabase(renderParams.MysqlDSN)
	if err != nil {
		return nil, err
	}

	//create configmap
	configMapCfg, err := f.Open("sophon-messager/sophon-messager-configmap.yaml")
	if err != nil {
		return nil, err
	}
	configMap, err := k8sEnv.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//deploy other node just service for others
	statefulSetCfg, err := f.Open("sophon-messager/sophon-messager-statefulset.yaml")
	if err != nil {
		return nil, err
	}
	statefulSet, err := k8sEnv.RunStatefulSets(ctx, func(ctx context.Context, k8sEnv *env.K8sEnvDeployer) ([]corev1.Pod, error) {
		return GetPods(ctx, k8sEnv, cfg.InstanceName)
	}, statefulSetCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create service
	svcCfg, err := f.Open("sophon-messager/sophon-messager-headless.yaml")
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
	return &SophonMessagerReturn{
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

func GetConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, configMapName string) (config.Config, error) {
	cfgData, err := k8sEnv.GetConfigMap(ctx, configMapName, "config.toml")
	if err != nil {
		return config.Config{}, err
	}

	var cfg config.Config
	err = toml.Unmarshal(cfgData, &cfg)
	if err != nil {
		return config.Config{}, err
	}

	return cfg, nil
}

func GetPods(ctx context.Context, k8sEnv *env.K8sEnvDeployer, instanceName string) ([]corev1.Pod, error) {
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("sophon-messager-%s-pod", env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), instanceName)))
}
