package chain_co

import (
	"context"
	"embed"
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

	AuthUrl string   `json:"-"`
	Nodes   []string `json:"-"`

	Replicas int `json:"replicas"`
}

type RenderParams struct {
	UniqueId string

	Config
}

func DefaultConfig() Config {
	return Config{Replicas: 0}
}

var PluginInfo = types.PluginInfo{
	Name:        string(env.ChainCo),
	Version:     version.Version(),
	Category:    types.Deploy,
	Description: "",
	Repo:        "https://github.com/ipfs-force-community/chain-co.git",
}

var _ env.IChainCoDeployer = (*ChainCoDeployer)(nil)

type ChainCoDeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	pods       []corev1.Pod
	deployment []*appv1.Deployment
	svc        *corev1.Service
}

func NewChainCoDeployer(env *env.K8sEnvDeployer, replicas int, authUrl string, ipEndpoints ...string) *ChainCoDeployer {
	return &ChainCoDeployer{
		env: env,
		cfg: &Config{
			Replicas: replicas, //default
			AuthUrl:  authUrl,
			Nodes:    ipEndpoints,
		},
	}
}

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params Config) (env.IDeployer, error) {
	cfg, err := utils.MergeStructAndInterface(DefaultConfig(), cfg, params)
	if err != nil {
		return nil, err
	}
	return &ChainCoDeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func (deployer *ChainCoDeployer) Name() string {
	return PluginInfo.Name
}

func (deployer *ChainCoDeployer) Pods() []corev1.Pod {
	return deployer.pods
}

func (deployer *ChainCoDeployer) Deployment() []*appv1.Deployment {
	return deployer.deployment
}

func (deployer *ChainCoDeployer) Svc() *corev1.Service {
	return deployer.svc
}

func (deployer *ChainCoDeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

//go:embed  chain-co
var f embed.FS

func (deployer *ChainCoDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		UniqueId: deployer.env.UniqueId(""),
		Config:   *deployer.cfg,
	}
	//create deployment
	deployCfg, err := f.Open("chain-co/chain-co-deployment.yaml")
	if err != nil {
		return err
	}

	deployment, err := deployer.env.RunDeployment(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.deployment = append(deployer.deployment, deployment)

	deployer.pods, err = deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-chain-co-%s-pod", deployer.env.UniqueId("")))
	if err != nil {
		return err
	}

	//create service
	svcCfg, err := f.Open("chain-co/chain-co-service.yaml")
	if err != nil {
		return err
	}
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
