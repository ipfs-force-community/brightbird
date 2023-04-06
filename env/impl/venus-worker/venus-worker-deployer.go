package venus_worker

import (
	"context"
	"embed"

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
	env.BaseRenderParams
	Config

	TestID string
}

func DefaultConfig() Config {
	return Config{}
}

var PluginInfo = types.PluginInfo{
	Name:        string(env.VenusWorker),
	Version:     version.Version(),
	Repo:        "https://github.com/ipfs-force-community/venus-cluster.git",
	ImageTarget: "venus-worker",
	Description: "",
}

var _ env.IVenusWorkerDeployer = (*VenusWorkerDeployer)(nil)

type VenusWorkerDeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	configMap  *corev1.ConfigMap
	pods       []corev1.Pod
	deployment []*appv1.Deployment
	svc        *corev1.Service
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

func (deployer *VenusWorkerDeployer) Pods() []corev1.Pod {
	return deployer.pods
}

func (deployer *VenusWorkerDeployer) Deployment() []*appv1.Deployment {
	return deployer.deployment
}

func (deployer *VenusWorkerDeployer) Svc() *corev1.Service {
	return deployer.svc
}

func (deployer *VenusWorkerDeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

var f embed.FS

func (deployer *VenusWorkerDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		BaseRenderParams: deployer.env.BaseRenderParams(),
		TestID:           deployer.env.TestID(),
		Config:           *deployer.cfg,
	}

	// create configMap
	configMap, err := f.Open("venus-worker/venus-worker-configmap.yaml")
	if err != nil {
		return err
	}
	deployer.configMap, err = deployer.env.CreateConfigMap(ctx, configMap, renderParams)
	if err != nil {
		return err
	}

	// create deployment
	deployCfg, err := f.Open("venus-worker/venus-worker-deployment.yaml")
	if err != nil {
		return err
	}
	deployment, err := deployer.env.RunDeployment(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.deployment = append(deployer.deployment, deployment)

	// create service
	svcCfg, err := f.Open("venus-worker/venus-worker-service.yaml")
	deployer.svc, err = deployer.env.RunService(ctx, svcCfg, renderParams)
	if err != nil {
		return err
	}

	deployer.svcEndpoint, err = deployer.env.WaitForServiceReady(ctx, deployer)
	if err != nil {
		return err
	}

	return nil
}
