package dropletmarket

import (
	"bytes"
	"context"
	"embed"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/ipfs-force-community/droplet/v2/config"
	logging "github.com/ipfs/go-log/v2"
	corev1 "k8s.io/api/core/v1"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	venusutils "github.com/ipfs-force-community/brightbird/pkg/env/venus_utils"
	types2 "github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"
)

var log = logging.Logger(PluginInfo.Name)

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

func GetPods(ctx context.Context, k8sEnv *env.K8sEnvDeployer, instanceName string) ([]corev1.Pod, error) {
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("droplet-market-%s-pod", env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), instanceName)))
}

func AddPieceStoragge(ctx context.Context, k8sEnv *env.K8sEnvDeployer, marketInstance DropletMarketDeployReturn, piecePvc, mountPath string) error {
	// 更新 configmap
	// 1. 得到新的config.toml文件
	// 2. 修改configmap的 config.toml 字段
	// 3. 更新configmap
	pods, err := GetPods(ctx, k8sEnv, marketInstance.InstanceName)
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return fmt.Errorf("not find pod for instancename : %v", marketInstance.InstanceName)
	}

	tomlBytes, err := k8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), "cat", "/root/.droplet/config.toml")
	if err != nil {
		return err
	}

	dropletCfg := *config.DefaultMarketConfig
	if err := toml.Unmarshal(tomlBytes, &dropletCfg); err != nil {
		return fmt.Errorf("unmarshal failed : %v ", err)
	}
	log.Debug("Unmarshal successed")

	for _, s := range dropletCfg.PieceStorage.Fs {
		if s.Name == piecePvc {
			log.Infof("PieceStorage.Fs.Name %v exist no need create", s.Name)
			return nil
		}
	}
	if err := dropletCfg.AddFsPieceStorage(&config.FsPieceStorage{
		Name:     piecePvc,
		ReadOnly: false,
		Path:     mountPath,
	}); err != nil {
		return fmt.Errorf("AddFsPieceStorage failed : %v ", err)
	}

	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(dropletCfg); err != nil {
		log.Infoln("Error encoding TOML: ", err)
		return err
	}
	tomlData := buf.String()
	log.Debugf("new config.toml is: \n%v", string(tomlData)) //nolint

	configmap, err := k8sEnv.GetConfigMapByName(ctx, marketInstance.ConfigMapName)
	if err != nil {
		return err
	}

	configmap.Data["config.toml"] = string(tomlData) //nolint

	err = k8sEnv.UpdateConfigMaps(ctx, configmap)
	if err != nil {
		return err
	}

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
		MountPath: mountPath,
	})

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

func GetPvcMountPath(ctx context.Context, k8sEnv *env.K8sEnvDeployer, clientInstance DropletMarketDeployReturn, piecePvc string) (string, error) {
	statefulset, err := k8sEnv.GetStatefulSet(ctx, clientInstance.StatefulSetName)
	if err != nil {
		return "", err
	}
	volumes := statefulset.Spec.Template.Spec.Volumes
	for _, vol := range volumes {
		if vol.Name == piecePvc {
			for _, cInfo := range statefulset.Spec.Template.Spec.Containers {
				for _, volm := range cInfo.VolumeMounts {
					if volm.Name == piecePvc {
						return volm.MountPath, nil
					}
				}
			}
		}
	}
	return "", fmt.Errorf("piece pvc %s exist", piecePvc)
}
