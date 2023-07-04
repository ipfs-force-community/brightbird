package sophongateway

import (
	"context"
	"embed"
	"fmt"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"github.com/ipfs-force-community/sophon-gateway/config"
	"github.com/naoina/toml"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	env.BaseConfig
	VConfig
}

type VConfig struct {
	AuthUrl    string `ignore:"-" json:"authUrl"`
	AdminToken string `ignore:"-" json:"adminToken"`

	Replicas int `json:"replicas" description:"number of replicas"`
}

type RenderParams struct {
	Config

	NameSpace       string
	PrivateRegistry string
	Args            []string
	UniqueId        string
}

type SophonGatewayReturn struct {
	VConfig
	env.CommonDeployParams
}

var PluginInfo = types.PluginInfo{
	Name:        "sophon-gateway",
	Version:     version.Version(),
	PluginType:  types.Deploy,
	Repo:        "https://github.com/ipfs-force-community/sophon-gateway.git",
	ImageTarget: "sophon-gateway",
	Description: "",
}

//go:embed sophon-gateway
var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, cfg Config) (*SophonGatewayReturn, error) {
	renderParams := RenderParams{
		NameSpace:       k8sEnv.NameSpace(),
		PrivateRegistry: k8sEnv.PrivateRegistry(),
		Args:            nil,
		UniqueId:        env.UniqueId(k8sEnv.TestID(), cfg.InstanceName),
		Config:          cfg,
	}
	//create configmap
	configMapCfg, err := f.Open("sophon-gateway/sophon-gateway-configmap.yaml")
	if err != nil {
		return nil, err
	}
	configMap, err := k8sEnv.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create deployment
	deployCfg, err := f.Open("sophon-gateway/sophon-gateway-statefulset.yaml")
	if err != nil {
		return nil, err
	}
	statefulSet, err := k8sEnv.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create service
	svcCfg, err := f.Open("sophon-gateway/sophon-gateway-headless.yaml")
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
	return &SophonGatewayReturn{
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

func Update(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params SophonGatewayReturn, updateCfg config.Config) error {
	cfgData, err := toml.Marshal(updateCfg)
	if err != nil {
		return err
	}
	err = k8sEnv.SetConfigMap(ctx, params.ConfigMapName, "config.toml", cfgData)
	if err != nil {
		return err
	}

	pods, err := GetPods(ctx, k8sEnv, params.InstanceName)
	if err != nil {
		return nil
	}

	for _, pod := range pods {
		_, err = k8sEnv.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.sophon-gateway/config.toml")
		if err != nil {
			return err
		}
	}

	return k8sEnv.UpdateStatefulSets(ctx, params.StatefulSetName)
}

func GetPods(ctx context.Context, k8sEnv *env.K8sEnvDeployer, instanceName string) ([]corev1.Pod, error) {
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("sophon-gateway-%s-pod", env.UniqueId(k8sEnv.TestID(), instanceName)))
}
