package dropletclient

import (
	"context"
	"embed"
	"fmt"

	venusutils "github.com/ipfs-force-community/brightbird/env/venus_utils"
	types2 "github.com/ipfs-force-community/brightbird/types"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/version"
	"github.com/ipfs-force-community/droplet/v2/config"
	"github.com/pelletier/go-toml/v2"
	corev1 "k8s.io/api/core/v1"
)

var log = logging.Logger("droplet-client-deployer")

type Config struct {
	env.BaseConfig
	VConfig
}

type VConfig struct {
	NodeUrl     string `jsonschema:"-" json:"nodeUrl"`
	WalletUrl   string `jsonschema:"-" json:"walletUrl"`
	WalletToken string `json:"walletToken" jsonschema:"walletToken" title:"WalletToken" description:"wallet用于鉴权的token" require:"true"`

	UserToken            string `json:"userToken" jsonschema:"userToken" title:"User Token" description:"user token" require:"true" `
	DefaultMarketAddress string `json:"defaultMarketAddress" jsonschema:"defaultMarketAddress" title:"DefaultMarketAddress" description:"当前droplet-client的默认地址" require:"true"`
}

type DropletClientDeployReturn struct { //nolint
	VConfig
	env.CommonDeployParams
	ClientToken string
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
		UniqueId:  env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), cfg.InstanceName),
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
	statefulSet, err := k8sEnv.RunStatefulSets(ctx, func(ctx context.Context, k8sEnv *env.K8sEnvDeployer) ([]corev1.Pod, error) {
		return GetPods(ctx, k8sEnv, cfg.InstanceName)
	}, deployCfg, renderParams)
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

	pods, err := GetPods(ctx, k8sEnv, cfg.InstanceName)
	if err != nil {
		return nil, fmt.Errorf("get pods fail %w", err)
	}

	clientToken, err := k8sEnv.ReadSmallFilelInPod(ctx, pods[0].GetName(), "/root/.droplet-client/token")
	if err != nil {
		return nil, fmt.Errorf("read token fail %w", err)
	}

	log.Infof("get droplet client token %s", string(clientToken))
	log.Debugln("statefulset is", statefulSet.GetName())

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
		ClientToken: string(clientToken),
	}, nil
}

func GetConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, configMapName string) (config.MarketClientConfig, error) {
	tomlBytes, err := k8sEnv.GetConfigMap(ctx, configMapName, "config.toml")
	if err != nil {
		return config.MarketClientConfig{}, err
	}
	log.Infoln("tomlBytes is: ", string(tomlBytes))

	var cfg config.MarketClientConfig
	err = toml.Unmarshal(tomlBytes, &cfg)
	if err != nil {
		log.Infoln("Unmarshal failed")
		return config.MarketClientConfig{}, err
	}
	log.Infoln("Unmarshal successed")

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
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("droplet-client-%s-pod", env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), instanceName)))
}

func AddPieceStoragge(ctx context.Context, k8sEnv *env.K8sEnvDeployer, clientInstance DropletClientDeployReturn, piecePvc, mountPath string) error {
	statefulset, err := k8sEnv.GetStatefulSet(ctx, clientInstance.StatefulSetName)
	if err != nil {
		return err
	}
	volumes := statefulset.Spec.Template.Spec.Volumes
	for _, vol := range volumes {
		if vol.Name == piecePvc {
			return fmt.Errorf("piece pvc %s exist", piecePvc)
		}
	}

	volumes = append(volumes, corev1.Volume{
		Name: piecePvc,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: piecePvc,
			},
		},
	})
	statefulset.Spec.Template.Spec.Volumes = volumes
	statefulset.Spec.Template.Spec.Containers[0].VolumeMounts = append(statefulset.Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
		Name:      piecePvc,
		MountPath: mountPath + piecePvc,
	})

	err = k8sEnv.UpdateStatefulSets(ctx, statefulset)
	if err != nil {
		return err
	}

	svc, err := k8sEnv.GetSvc(ctx, clientInstance.SVCName)
	if err != nil {
		return err
	}
	
	_, err = k8sEnv.WaitForServiceReady(ctx, svc, venusutils.VenusHealthCheck)
	if err != nil {
		return err
	}

	return nil
}
