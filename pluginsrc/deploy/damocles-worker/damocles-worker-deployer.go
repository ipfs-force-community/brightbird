package damoclesworker

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type DamoclesWorkerConfig string //nolint

type Config struct {
	env.BaseConfig
	DamoclesManagerURL string `json:"-"`
	AuthToken          string `json:"-"`

	MinerAddress address.Address `json:"-"`
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
	Name:        "damocles-worker",
	Version:     version.Version(),
	PluginType:  types.Deploy,
	Repo:        "https://github.com/ipfs-force-community/damocles.git",
	ImageTarget: "damocles-worker",
	Description: "",
}

var _ env.IDeployer = (*DamoclesWorkerDeployer)(nil)

type DamoclesWorkerDeployer struct { //nolint
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
	return &DamoclesWorkerDeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func (deployer *DamoclesWorkerDeployer) InstanceName() (string, error) {
	return deployer.cfg.InstanceName, nil
}

func (deployer *DamoclesWorkerDeployer) Pods(ctx context.Context) ([]corev1.Pod, error) {
	return deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("damocles-worker-%s-pod", env.UniqueId(deployer.env.TestID(), deployer.cfg.InstanceName)))
}

func (deployer *DamoclesWorkerDeployer) StatefulSet(ctx context.Context) (*appv1.StatefulSet, error) {
	return deployer.env.GetStatefulSet(ctx, deployer.statefulSetName)
}

func (deployer *DamoclesWorkerDeployer) Svc(ctx context.Context) (*corev1.Service, error) {
	return deployer.env.GetSvc(ctx, deployer.svcName)
}

func (deployer *DamoclesWorkerDeployer) SvcEndpoint() (types.Endpoint, error) {
	return deployer.svcEndpoint, nil
}

func (deployer *DamoclesWorkerDeployer) Param(key string) (env.Params, error) {
	return env.Params{}, errors.New("no params")
}

var f embed.FS

func (deployer *DamoclesWorkerDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		NameSpace:       deployer.env.NameSpace(),
		PrivateRegistry: deployer.env.PrivateRegistry(),
		TestID:          deployer.env.TestID(),
		Config:          *deployer.cfg,
	}

	// create configMap
	configMapCfg, err := f.Open("damocles-worker/damocles-worker-configmap.yaml")
	if err != nil {
		return err
	}
	configMap, err := deployer.env.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.configMapName = configMap.GetName()

	// create deployment
	deployCfg, err := f.Open("damocles-worker/damocles-worker-statefulset.yaml")
	if err != nil {
		return err
	}
	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.statefulSetName = statefulSet.GetName()

	// create service
	svcCfg, err := f.Open("damocles-worker/damocles-worker-headless.yaml")
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

func (deployer *DamoclesWorkerDeployer) GetConfig(ctx context.Context) (env.Params, error) {
	cfgData, err := deployer.env.GetConfigMap(ctx, deployer.configMapName, "damocles-worker.toml")
	if err != nil {
		return env.Params{}, err
	}

	return env.ParamsFromVal(cfgData), nil
}

func (deployer *DamoclesWorkerDeployer) Update(ctx context.Context, updateCfg interface{}) error {
	if updateCfg != nil {
		update := updateCfg.(*DamoclesWorkerConfig)
		err := deployer.env.SetConfigMap(ctx, deployer.configMapName, "damocles-worker.toml", []byte(*update))
		if err != nil {
			return err
		}
		pods, err := deployer.Pods(ctx)
		if err != nil {
			return nil
		}
		for _, pod := range pods {
			_, err = deployer.env.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(*update)+"'", ">", "/root/.damocles-worker/damocles-worker.toml")
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
