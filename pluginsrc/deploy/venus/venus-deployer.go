package venus

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"

	vCfg "github.com/filecoin-project/venus/pkg/config"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	env.BaseConfig
	VConfig
}

type VConfig struct {
	AuthUrl        string   `ignore:"-" json:"authUrl"`
	AdminToken     string   `ignore:"-" json:"adminToken"`
	BootstrapPeers []string `ignore:"-" json:"bootstrapPeers"`

	NetType  string `json:"netType" description:"network type: mainnet,2k,calibrationnet,force"`
	Replicas int    `json:"replicas" description:"number of replicas"`
}

type RenderParams struct {
	Config

	NameSpace       string
	PrivateRegistry string
	Args            []string

	UniqueId string
}

type VenusDeployReturn struct {
	VConfig
	env.CommonDeployParams
}

func DefaultConfig() Config {
	return Config{
		VConfig: VConfig{
			Replicas: 1,
			NetType:  "force",
		},
	}
}

var PluginInfo = types.PluginInfo{
	Name:        "venus-daemon",
	Version:     version.Version(),
	PluginType:  types.Deploy,
	Repo:        "https://github.com/filecoin-project/venus.git",
	ImageTarget: "venus",
	Description: "",
}

//go:embed venus-node
var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, incomineCfg Config) (*VenusDeployReturn, error) {
	renderParams := RenderParams{
		NameSpace:       k8sEnv.NameSpace(),
		PrivateRegistry: k8sEnv.PrivateRegistry(),
		UniqueId:        env.UniqueId(k8sEnv.TestID(), incomineCfg.InstanceName),
		Args:            buildArgs(incomineCfg.BootstrapPeers, incomineCfg.NetType),
		Config:          incomineCfg,
	}

	//create configmap
	configMapCfg, err := f.Open("venus-node/venus-configmap.yaml")
	if err != nil {
		return nil, err
	}

	configMap, err := k8sEnv.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create statefulset
	deployCfg, err := f.Open("venus-node/venus-node-stateful-deployment.yaml")
	if err != nil {
		return nil, err
	}
	statefulSet, err := k8sEnv.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create headless service
	svcCfg, err := f.Open("venus-node/venus-node-headless.yaml")
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
	return &VenusDeployReturn{
		VConfig: incomineCfg.VConfig,
		CommonDeployParams: env.CommonDeployParams{
			BaseConfig:      incomineCfg.BaseConfig,
			DeployName:      PluginInfo.Name,
			StatefulSetName: statefulSet.GetName(),
			ConfigMapName:   configMap.GetName(),
			SVCName:         svc.GetName(),
			SvcEndpoint:     svcEndpoint,
		},
	}, nil
}

func GetConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, configMapName string) (vCfg.Config, error) {
	cfgData, err := k8sEnv.GetConfigMap(ctx, configMapName, "config.json")
	if err != nil {
		return vCfg.Config{}, err
	}

	cfg := vCfg.Config{}
	err = json.Unmarshal(cfgData, &cfg)
	if err != nil {
		return vCfg.Config{}, err
	}

	return cfg, nil
}

// Update
// todo change this mode to config
func Update(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params VenusDeployReturn, updateCfg interface{}) error {
	if updateCfg != nil {
		cfgData, err := json.Marshal(updateCfg)
		if err != nil {
			return err
		}
		err = k8sEnv.SetConfigMap(ctx, params.ConfigMapName, "config.json", cfgData)
		if err != nil {
			return err
		}

		pods, err := GetPods(ctx, k8sEnv, params.InstanceName)
		if err != nil {
			return nil
		}
		for _, pod := range pods {
			_, err = k8sEnv.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.venus/config.json")
			if err != nil {
				return err
			}
		}
	}

	err := k8sEnv.UpdateStatefulSets(ctx, params.StatefulSetName)
	if err != nil {
		return err
	}
	return nil
}

func buildArgs(bootstrapPeers []string, netType string) []string {
	args := []string{
		"daemon",
		"--genesisfile=/shared-dir/devgen.car",
		"--import-snapshot=/shared-dir/dev-snapshot.car",
		"--network=" + netType,
	}
	return args
}

func GetPods(ctx context.Context, k8sEnv *env.K8sEnvDeployer, instanceName string) ([]corev1.Pod, error) {

	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("venus-%s-pod", env.UniqueId(k8sEnv.TestID(), instanceName)))
}
