package venus_worker

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	env.BaseConfig
	VenusSectorManagerUrl string `json:"-"`
	AuthToken             string `json:"-"`

	MinerAddress string `json:"minerAddress"`
}

type RenderParams struct {
	Config

	NameSpace       string
	PrivateRegistry string
	Args            []string

	TestID string
}

func DefaultConfig() Config {
	return Config{}
}

var PluginInfo = types.PluginInfo{
	Name:        "venus-worker",
	Version:     version.Version(),
	Repo:        "https://github.com/ipfs-force-community/venus-cluster.git",
	ImageTarget: "venus-worker",
	Description: "",
}

var _ env.IDeployer = (*VenusWorkerDeployer)(nil)

type VenusWorkerDeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	configMapName   string
	statefulSetName string
	svcName         string
}

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params Config) (env.IDeployer, error) {
	cfg, err := utils.MergeStructAndInterface(DefaultConfig(), cfg, params)
	if err != nil {
		return nil, err
	}
	return &VenusWorkerDeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func (deployer *VenusWorkerDeployer) Name() string {
	return PluginInfo.Name
}

func (deployer *VenusWorkerDeployer) Pods(ctx context.Context) ([]corev1.Pod, error) {
	return deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-worker-%s-pod", deployer.env.UniqueId(deployer.cfg.SvcMap[types.OutLabel])))
}

func (deployer *VenusWorkerDeployer) StatefulSet(ctx context.Context) (*appv1.StatefulSet, error) {
	return deployer.env.GetStatefulSet(ctx, deployer.statefulSetName)
}

func (deployer *VenusWorkerDeployer) Svc(ctx context.Context) (*corev1.Service, error) {
	return deployer.env.GetSvc(ctx, deployer.svcName)
}

func (deployer *VenusWorkerDeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

func (deployer *VenusWorkerDeployer) Param(key string) (interface{}, error) {
	return nil, errors.New("no params")
}

var f embed.FS

func (deployer *VenusWorkerDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		NameSpace:       deployer.env.NameSpace(),
		PrivateRegistry: deployer.env.PrivateRegistry(),
		TestID:          deployer.env.TestID(),
		Config:          *deployer.cfg,
	}

	// create configMap
	configMapCfg, err := f.Open("venus-worker/venus-worker-configmap.yaml")
	if err != nil {
		return err
	}
	configMap, err := deployer.env.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.configMapName = configMap.GetName()

	// create deployment
	deployCfg, err := f.Open("venus-worker/venus-worker-statefulset.yaml")
	if err != nil {
		return err
	}
	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.statefulSetName = statefulSet.GetName()

	// create service
	svcCfg, err := f.Open("venus-worker/venus-worker-headless.yaml")
	if err != nil {
		return err
	}

	svc, err := deployer.env.RunService(ctx, svcCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.svcName = svc.GetName()

	deployer.svcEndpoint, err = deployer.env.WaitForServiceReady(ctx, deployer)
	if err != nil {
		return err
	}

	return nil
}

func (deployer *VenusWorkerDeployer) GetConfig(ctx context.Context) (interface{}, error) {
	cfgData, err := deployer.env.GetConfigMap(ctx, deployer.configMapName, "venus-worker.toml")
	if err != nil {
		return nil, err
	}

	return (*env.VenusWorkerConfig)(utils.StringPtr(string(cfgData))), nil
}

func (deployer *VenusWorkerDeployer) Update(ctx context.Context, updateCfg interface{}) error {
	if updateCfg != nil {
		update := updateCfg.(*env.VenusWorkerConfig)
		err := deployer.env.SetConfigMap(ctx, deployer.configMapName, "venus-worker.toml", []byte(*update))
		if err != nil {
			return err
		}
		pods, err := deployer.Pods(ctx)
		if err != nil {
			return nil
		}
		for _, pod := range pods {
			_, err = deployer.env.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(*update)+"'", ">", "/root/.venus-worker/venus-worker.toml")
			if err != nil {
				return err
			}
		}
	}

	err := deployer.env.UpdateStatefulSets(ctx, deployer.statefulSetName)
	if err != nil {
		return err
	}
	return nil
}
