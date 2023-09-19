package dropletclient

import (
	"context"
	"embed"
	"fmt"

	venusutils "github.com/ipfs-force-community/brightbird/env/venus_utils"
	types2 "github.com/ipfs-force-community/brightbird/types"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/version"
	"github.com/ipfs-force-community/droplet/v2/config"
	"github.com/pelletier/go-toml/v2"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	env.BaseConfig
	VConfig
}

type VConfig struct {
	NodeUrl     string `jsonschema:"-" json:"nodeUrl"`
	WalletUrl   string `jsonschema:"-" json:"walletUrl"`
	WalletToken string `jsonschema:"-" json:"walletToken"`

	UserToken  string `json:"userToken" jsonschema:"userToken" title:"User Token" description:"user token" require:"true" `
	ClientAddr string `json:"clientAddr" jsonschema:"clientAddr" title:"Client Address" description:"pay for storage/retrieval" require:"true" `
}

type DropletClientDeployReturn struct { //nolint
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

func DefaultConfig() Config {
	return Config{}
}

var PluginInfo = types2.PluginInfo{
	Name:        "droplet-client",
	Version:     version.Version(),
	PluginType:  types2.Deploy,
	Description: "",
	DeployPluginParams: types2.DeployPluginParams{
		Repo:        "https://github.com/ipfs-force-community/droplet.git",
		ImageTarget: "droplet-client",
		BuildScript: `make docker-push TAG={{.Commit}} BUILD_DOCKER_PROXY={{.Proxy}} PRIVATE_REGISTRY={{.Registry}}`,
	},
}

//go:embed  droplet-client
var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, cfg Config) (*DropletClientDeployReturn, error) {
	renderParams := RenderParams{
		NameSpace: k8sEnv.NameSpace(),
		Registry:  k8sEnv.Registry(),
		Args:      nil,
		UniqueId:  env.UniqueId(k8sEnv.TestID(), cfg.InstanceName),
		Config:    cfg,
	}
	//create configmap
	configMapCfg, err := f.Open("droplet-client/droplet-client-configmap.yaml")
	if err != nil {
		return nil, err
	}
	configMap, err := k8sEnv.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create deployment
	deployCfg, err := f.Open("droplet-client/droplet-client-statefulset.yaml")
	if err != nil {
		return nil, err
	}
	statefulSet, err := k8sEnv.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create service
	svcCfg, err := f.Open("droplet-client/droplet-client-headless.yaml")
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
	return &DropletClientDeployReturn{
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

func GetConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, configMapName string) (config.MarketClientConfig, error) {
	cfgData, err := k8sEnv.GetConfigMap(ctx, configMapName, "config.toml")
	if err != nil {
		return config.MarketClientConfig{}, err
	}

	var cfg config.MarketClientConfig
	err = toml.Unmarshal(cfgData, &cfg)
	if err != nil {
		return config.MarketClientConfig{}, err
	}

	return cfg, nil
}

func Update(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params DropletClientDeployReturn, updateCfg config.MarketClientConfig) error {
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
		_, err = k8sEnv.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.droplet-client/config.toml")
		if err != nil {
			return err
		}
	}

	return k8sEnv.UpdateStatefulSetsByName(ctx, params.StatefulSetName)
}

func GetPods(ctx context.Context, k8sEnv *env.K8sEnvDeployer, instanceName string) ([]corev1.Pod, error) {
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("droplet-client-%s-pod", env.UniqueId(k8sEnv.TestID(), instanceName)))
}
