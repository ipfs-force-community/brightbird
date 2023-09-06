package damoclesmanager

import (
	"context"
	"embed"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/ipfs-force-community/brightbird/env"
	venusutils "github.com/ipfs-force-community/brightbird/env/venus_utils"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
	"github.com/pelletier/go-toml"
	corev1 "k8s.io/api/core/v1"
)

// Config 定义了配置结构体
type Config struct {
	env.BaseConfig
	VConfig
}

// VConfig 定义了具体配置的结构体
type VConfig struct {
	{{ConfigName1}} address.Address `json:"{{ConfigName1}}"  jsonschema:"{{ConfigName1}}" title:"{{ConfigName1}}" require:"true" `
	{{ConfigName2}} string          `jsonschema:"-"`
}

// DamoclesManagerReturn 定义了返回结构体
type DamoclesManagerReturn struct {
	VConfig
	env.CommonDeployParams
}

// RenderParams 定义了渲染参数结构体
type RenderParams struct {
	Config

	NameSpace       string
	Registry string
	Args            []string

	UniqueId string
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{}
}

// PluginInfo 定义了插件信息
var PluginInfo = types.PluginInfo{
	Name:        "{{plugin-name}}",
	Version:     version.Version(),
	PluginType:  types.Deploy,
	DeployPluginParams: types.DeployPluginParams{
		Repo:        "{{plugin-github-url}}",
		ImageTarget: "{{plugin-name}}",
	},
	Description: "{{plugin-description}}",
}

// f 为嵌入资源文件系统
var f embed.FS

// DeployFromConfig 从配置中部署
func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, cfg Config) (*DamoclesManagerReturn, error) {
	renderParams := RenderParams{
		NameSpace:       k8sEnv.NameSpace(),
		Registry: 		 k8sEnv.Registry(),
		UniqueId:        env.UniqueId(k8sEnv.TestID(), cfg.InstanceName),
		Config:          cfg,
	}

	// 创建 ConfigMap
	configMapFs, err := f.Open("{{plugin-name}}/{{plugin-name}}-configmap.yaml")
	if err != nil {
		return nil, err
	}
	configMap, err := k8sEnv.RunConfigMap(ctx, configMapFs, renderParams)
	if err != nil {
		return nil, err
	}

	// 创建 Deployment
	deployCfg, err := f.Open("{{plugin-name}}/{{plugin-name}}-statefulset.yaml")
	if err != nil {
		return nil, err
	}
	statefulSet, err := k8sEnv.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return nil, err
	}

	// 创建 Service
	svcCfg, err := f.Open("{{plugin-name}}/{{plugin-name}}-headless.yaml")
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

	return &DamoclesManagerReturn{
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

// GetConfig 获取配置
func GetConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, configMapName string) (interface{}, error) {
	cfgData, err := k8sEnv.GetConfigMap(ctx, configMapName, "{{config-file-name}}")
	if err != nil {
		return nil, err
	}

	var cfg interface{}
	err = toml.Unmarshal(cfgData, &cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Update 更新配置
func Update(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params DamoclesManagerReturn, updateCfg interface{}) error {
	cfgData, err := toml.Marshal(updateCfg)
	if err != nil {
		return err
	}
	err = k8sEnv.SetConfigMap(ctx, params.ConfigMapName, "{{config-file-name}}", cfgData)
	if err != nil {
		return err
	}

	pods, err := GetPods(ctx, k8sEnv, params.InstanceName)
	if err != nil {
		return nil
	}
	for _, pod := range pods {
		_, err = k8sEnv.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.{{plugin-name}}/{{config-file-name}}")
		if err != nil {
			return err
		}
	}

	err = k8sEnv.UpdateStatefulSets(ctx, params.StatefulSetName)
	if err != nil {
		return err
	}
	return nil
}

// GetPods 获取 Pod 列表
func GetPods(ctx context.Context, k8sEnv *env.K8sEnvDeployer, instanceName string) ([]corev1.Pod, error) {
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("{{plugin-name}}-%s-pod", env.UniqueId(k8sEnv.TestID(), instanceName)))
}