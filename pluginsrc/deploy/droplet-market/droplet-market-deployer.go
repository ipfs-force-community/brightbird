package dropletmarket

import (
	"context"
	"embed"
	"fmt"

	"github.com/ipfs-force-community/brightbird/env"
	venusutils "github.com/ipfs-force-community/brightbird/env/venus_utils"
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

	PieceStores []string `jsonschema:"-"`
	NodeUrl     string   `jsonschema:"-"`
	GatewayUrl  string   `jsonschema:"-"`
	MessagerUrl string   `jsonschema:"-"`
	AuthUrl     string   `jsonschema:"-"`
}

type DropletMarketDeployReturn struct { //nolint
	VConfig
	env.CommonDeployParams
}

type RenderParams struct {
	Config
	NameSpace string
	Registry  string
	UniqueId  string
	MysqlDSN  string
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
		UniqueId:  env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), cfg.InstanceName),
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
	statefulSet, err := k8sEnv.RunStatefulSets(ctx, func(ctx context.Context, k8sEnv *env.K8sEnvDeployer) ([]corev1.Pod, error) {
		return GetPods(ctx, k8sEnv, cfg.InstanceName)
	}, deployCfg, renderParams)
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

	svcEndpoint, err := k8sEnv.WaitForServiceReady(ctx, svc, venusutils.VenusHealthCheck)
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

func GetPods(ctx context.Context, k8sEnv *env.K8sEnvDeployer, instanceName string) ([]corev1.Pod, error) {
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("droplet-market-%s-pod", env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), instanceName)))
}

func AddPieceStoragge(ctx context.Context, k8sEnv *env.K8sEnvDeployer, marketInstance DropletMarketDeployReturn, piecePvc, mountPath string) error {
	statefulset, err := k8sEnv.GetStatefulSet(ctx, marketInstance.StatefulSetName)
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
	//restart
	err = k8sEnv.UpdateStatefulSets(ctx, statefulset)
	if err != nil {
		return err
	}

	svc, err := k8sEnv.GetSvc(ctx, marketInstance.SVCName)
	if err != nil {
		return err
	}
	_, err = k8sEnv.WaitForServiceReady(ctx, svc, venusutils.VenusHealthCheck)
	if err != nil {
		return err
	}

	return nil
}
