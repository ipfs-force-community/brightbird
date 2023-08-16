package sophonminer

import (
	"context"
	"embed"
	"fmt"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
	"github.com/ipfs-force-community/sophon-miner/node/config"
	"github.com/pelletier/go-toml"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	env.BaseConfig
	VConfig
}

type VConfig struct {
	NodeUrl    string `jsonschema:"-" json:"nodeUrl"`
	AuthUrl    string `jsonschema:"-" json:"authUrl"`
	GatewayUrl string `jsonschema:"-" json:"gatewayUrl"`
	AuthToken  string `jsonschema:"-" json:"authToken"`

	UseMysql bool `json:"useMysql" jsonschema:"useMysql" title:"UserMysql" require:"true" description:"true or false"`
}

type SophonMinerDeployReturn struct { //nolint
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

var PluginInfo = types.PluginInfo{
	Name:       "sophon-miner",
	Version:    version.Version(),
	PluginType: types.Deploy,
	DeployPluginParams: types.DeployPluginParams{
		Repo:        "https://github.com/ipfs-force-community/sophon-miner.git",
		ImageTarget: "sophon-miner",
		BuildScript: `make docker-push TAG={{.Commit}} BUILD_DOCKER_PROXY={{.Proxy}} PRIVATE_REGISTRY={{.Registry}}`,
	},
	Description: "",
}

//go:embed sophon-miner
var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, cfg Config) (*SophonMinerDeployReturn, error) {
	renderParams := RenderParams{
		NameSpace: k8sEnv.NameSpace(),
		Registry:  k8sEnv.Registry(),
		UniqueId:  env.UniqueId(k8sEnv.TestID(), cfg.InstanceName),
		Config:    cfg,
	}
	if cfg.UseMysql {
		renderParams.MysqlDSN = k8sEnv.FormatMysqlConnection("sophon-miner-" + renderParams.UniqueId)
	}
	if len(renderParams.MysqlDSN) > 0 {
		err := k8sEnv.ResourceMgr().EnsureDatabase(renderParams.MysqlDSN)
		if err != nil {
			return nil, err
		}
	}
	//create configmap
	configMapCfg, err := f.Open("sophon-miner/sophon-miner-configmap.yaml")
	if err != nil {
		return nil, err
	}
	configMap, err := k8sEnv.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create deployment
	deployCfg, err := f.Open("sophon-miner/sophon-miner-statefulset.yaml")
	if err != nil {
		return nil, err
	}
	statefulSet, err := k8sEnv.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create service
	svcCfg, err := f.Open("sophon-miner/sophon-miner-headless.yaml")
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

	return &SophonMinerDeployReturn{
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

func GetConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, configMapName string) (config.MinerConfig, error) {
	cfgData, err := k8sEnv.GetConfigMap(ctx, configMapName, "config.toml")
	if err != nil {
		return config.MinerConfig{}, err
	}

	var cfg config.MinerConfig
	err = toml.Unmarshal(cfgData, &cfg)
	if err != nil {
		return config.MinerConfig{}, err
	}
	return cfg, nil
}

// Update
// todo change this mode to config
func Update(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params SophonMinerDeployReturn, updateCfg config.MinerConfig) error {
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
		_, err = k8sEnv.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.sophon-miner/config.toml")
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
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("sophon-miner-%s-pod", env.UniqueId(k8sEnv.TestID(), instanceName)))
}
