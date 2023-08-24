package damoclesmanager

import (
	"context"
	"embed"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
	"github.com/pelletier/go-toml"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	env.BaseConfig
	VConfig
}

type VConfig struct {
	NodeUrl     	string `jsonschema:"-"`
	MessagerUrl 	string `jsonschema:"-"`
	MarketUrl   	string `jsonschema:"-"`
	GatewayUrl  	string `jsonschema:"-"`
	AuthUrl     	string `jsonschema:"-"`
	MinerAddress    string `jsonschema:"-"`

	SenderWalletAddress address.Address `json:"senderWalletAddress"  jsonschema:"senderWalletAddress" title:"SenderWalletAddress" require:"true" `
	UserToken           string          `json:"userToken" jsonschema:"userToken" title:"UserToken" require:"true" `
}

type DamoclesManagerReturn struct { //nolint
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
	Name:       "damocles-manager",
	Version:    version.Version(),
	PluginType: types.Deploy,
	DeployPluginParams: types.DeployPluginParams{
		Repo:        "https://github.com/ipfs-force-community/damocles.git",
		ImageTarget: "damocles-manager",
		BuildScript: `make docker-push TAG={{.Commit}} BUILD_DOCKER_PROXY={{.Proxy}} PRIVATE_REGISTRY={{.Registry}}`,
	},
	Description: "",
}

//go:embed damocles-manager
var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, cfg Config) (*DamoclesManagerReturn, error) {
	renderParams := RenderParams{
		NameSpace: k8sEnv.NameSpace(),
		Registry:  k8sEnv.Registry(),
		UniqueId:  env.UniqueId(k8sEnv.TestID(), cfg.InstanceName),
		Config:    cfg,
	}

	// create configMap
	configMapFs, err := f.Open("damocles-manager/damocles-manager-configmap.yaml")
	if err != nil {
		return nil, err
	}
	configMap, err := k8sEnv.RunConfigMap(ctx, configMapFs, renderParams)
	if err != nil {
		return nil, err
	}

	// create deployment
	deployCfg, err := f.Open("damocles-manager/damocles-manager-statefulset.yaml")
	if err != nil {
		return nil, err
	}
	statefulSet, err := k8sEnv.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return nil, err
	}

	// create service
	svcCfg, err := f.Open("damocles-manager/damocles-manager-headless.yaml")
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

func GetConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, configMapName string) (interface{}, error) {
	cfgData, err := k8sEnv.GetConfigMap(ctx, configMapName, "sector-manager.cfg")
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

func Update(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params DamoclesManagerReturn, updateCfg interface{}) error {
	cfgData, err := toml.Marshal(updateCfg)
	if err != nil {
		return err
	}
	err = k8sEnv.SetConfigMap(ctx, params.ConfigMapName, "sector-manager.cfg", cfgData)
	if err != nil {
		return err
	}

	pods, err := GetPods(ctx, k8sEnv, params.InstanceName)
	if err != nil {
		return nil
	}
	for _, pod := range pods {
		_, err = k8sEnv.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.damocles-manager/sector-manager.cfg")
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

func GetPods(ctx context.Context, k8sEnv *env.K8sEnvDeployer, instanceName string) ([]corev1.Pod, error) {
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("damocles-manager-%s-pod", env.UniqueId(k8sEnv.TestID(), instanceName)))
}
