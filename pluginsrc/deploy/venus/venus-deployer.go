package venus

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"

	vCfg "github.com/filecoin-project/venus/pkg/config"
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
	AuthUrl        string   `jsonschema:"-" json:"authUrl"`
	AdminToken     string   `jsonschema:"-" json:"adminToken"`
	BootstrapPeers []string `jsonschema:"-" json:"bootstrapPeers"`

	NetType  string `json:"netType" jsonschema:"netType" title:"Network Type" default:"force" require:"true" description:"network type: mainnet,2k,calibrationnet,force" enum:"mainnet,2k,calibrationnet,force"`
	Replicas int    `json:"replicas"  jsonschema:"replicas" title:"Replicas" default:"1" require:"true" description:"number of replicas"`

	GenesisStorage  string `json:"genesisStorage"  jsonschema:"genesisStorage" title:"GenesisStorage" default:"" require:"true" description:"used genesis file"`
	SnapshotStorage string `json:"snapshotStorage"  jsonschema:"snapshotStorage" title:"SnapshotStorage" default:"" require:"true" description:"used to read snapshot file"`
}

type RenderParams struct {
	Config

	NameSpace string
	Registry  string
	Args      []string

	UniqueId string
}

type VenusDeployReturn struct { //nolint
	VConfig
	env.CommonDeployParams
}

var PluginInfo = types.PluginInfo{
	Name:       "venus-daemon",
	Version:    version.Version(),
	PluginType: types.Deploy,
	DeployPluginParams: types.DeployPluginParams{
		Repo:        "https://github.com/filecoin-project/venus.git",
		ImageTarget: "venus",
		BuildScript: `make docker-push TAG={{.Commit}} BUILD_DOCKER_PROXY={{.Proxy}} PRIVATE_REGISTRY={{.Registry}}`,
	},
	Description: "",
}

//go:embed venus-node
var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, incomineCfg Config) (*VenusDeployReturn, error) {
	renderParams := RenderParams{
		NameSpace: k8sEnv.NameSpace(),
		Registry:  k8sEnv.Registry(),
		UniqueId:  env.UniqueId(k8sEnv.TestID(), incomineCfg.InstanceName),
		Config:    incomineCfg,
	}

	args := []string{
		"daemon",
		"--network=" + incomineCfg.NetType,
	}

	if len(incomineCfg.GenesisStorage) > 0 {
		args = append(args, "--genesisfile=/root/devgen/devgen.car")
	}

	if len(incomineCfg.SnapshotStorage) > 0 {
		args = append(args, "--import-snapshot=/root/snapshop/snapshot.car")
	}
	renderParams.Args = args

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
	statefulSet, err := k8sEnv.RunStatefulSets(ctx, func(ctx context.Context, k8sEnv *env.K8sEnvDeployer) ([]corev1.Pod, error) {
		return GetPods(ctx, k8sEnv, incomineCfg.InstanceName)
	}, deployCfg, renderParams)
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

	svcEndpoint, err := k8sEnv.WaitForServiceReady(ctx, svc, venusutils.VenusHealthCheck)
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

	err := k8sEnv.UpdateStatefulSetsByName(ctx, params.StatefulSetName)
	if err != nil {
		return err
	}
	return nil
}

func GetPods(ctx context.Context, k8sEnv *env.K8sEnvDeployer, instanceName string) ([]corev1.Pod, error) {
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("venus-%s-pod", env.UniqueId(k8sEnv.TestID(), instanceName)))
}
