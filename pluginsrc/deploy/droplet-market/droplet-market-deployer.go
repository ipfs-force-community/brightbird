package dropletmarket

import (
	"context"
	"embed"
	"fmt"

	"github.com/ipfs-force-community/brightbird/env"
	types2 "github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
	"github.com/ipfs-force-community/droplet/v2/config"
	"github.com/pelletier/go-toml"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	env.BaseConfig
	VConfig
}

type VConfig struct {
	UserToken string `json:"userToken" jsonschema:"userToken" title:"UserToken" require:"true"`
	UseMysql  bool   `json:"useMysql" jsonschema:"useMysql" title:"UserMysql" require:"true" description:"true or false"`

	NodeUrl     string `jsonschema:"-"`
	GatewayUrl  string `jsonschema:"-"`
	MessagerUrl string `jsonschema:"-"`
	AuthUrl     string `jsonschema:"-"`
}

type DropletMarketDeployReturn struct { //nolint
	VConfig
	env.CommonDeployParams
}

type RenderParams struct {
	Config

	NameSpace string
	Registry  string
	Args      []string

	UniqueId string
	MysqlDSN string
}

func DefaultConfig() Config {
	return Config{}
}

var PluginInfo = types2.PluginInfo{
	Name:       "droplet",
	Version:    version.Version(),
	PluginType: types2.Deploy,
	DeployPluginParams: types2.DeployPluginParams{
		Repo:        "https://github.com/ipfs-force-community/droplet.git",
		ImageTarget: "droplet",
		BuildScript: `make docker-push TAG={{.Commit}} BUILD_DOCKER_PROXY={{.Proxy}} PRIVATE_REGISTRY={{.Registry}}`,
	},
	Description: "",
}

//go:embed droplet-market
var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, cfg Config) (*DropletMarketDeployReturn, error) {
	renderParams := RenderParams{
		NameSpace: k8sEnv.NameSpace(),
		Registry:  k8sEnv.Registry(),
		UniqueId:  env.UniqueId(k8sEnv.TestID(), cfg.InstanceName),
		Config:    cfg,
	}
	if cfg.UseMysql {
		renderParams.MysqlDSN = k8sEnv.FormatMysqlConnection("droplet-market-" + renderParams.UniqueId)
		err := k8sEnv.ResourceMgr().EnsureDatabase(renderParams.MysqlDSN)
		if err != nil {
			return nil, err
		}
	}
	//create configmap
	configMapCfg, err := f.Open("droplet-market/droplet-market-configmap.yaml")
	if err != nil {
		return nil, err
	}
	configMap, err := k8sEnv.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create deployment
	deployCfg, err := f.Open("droplet-market/droplet-market-statefulset.yaml")
	if err != nil {
		return nil, err
	}
	statefulSet, err := k8sEnv.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create service
	svcCfg, err := f.Open("droplet-market/droplet-market-headless.yaml")
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

func GetConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, configMapName string) (config.MarketConfig, error) {
	cfgData, err := k8sEnv.GetConfigMap(ctx, configMapName, "config.toml")
	if err != nil {
		return config.MarketConfig{}, err
	}

	var cfg config.MarketConfig
	err = toml.Unmarshal(cfgData, &cfg)
	if err != nil {
		return config.MarketConfig{}, err
	}

	return cfg, nil
}

func Update(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params DropletMarketDeployReturn, updateCfg config.MarketConfig) error {
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
		_, err = k8sEnv.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.droplet-market/config.toml")
		if err != nil {
			return err
		}
	}

	return k8sEnv.UpdateStatefulSets(ctx, params.StatefulSetName)
}

func GetPods(ctx context.Context, k8sEnv *env.K8sEnvDeployer, instanceName string) ([]corev1.Pod, error) {
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("droplet-market-%s-pod", env.UniqueId(k8sEnv.TestID(), instanceName)))
}
