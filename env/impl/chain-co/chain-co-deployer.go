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

	AuthUrl    string   `json:"-"`
	AdminToken string   `json:"-"`
	Nodes      []string `json:"-"`

	Replicas int `json:"replicas"`
}

type RenderParams struct {
	env.BaseRenderParams
	Config

	UniqueId string
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
	ImageTarget: "chain-co",
}

var _ env.IChainCoDeployer = (*ChainCoDeployer)(nil)

type ChainCoDeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	pods         []corev1.Pod
	statefulSets []*appv1.StatefulSet
	svc          *corev1.Service
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
	return nil
}

func (deployer *ChainCoDeployer) StatefulSets() []*appv1.StatefulSet {
	return deployer.statefulSets
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
		BaseRenderParams: deployer.env.BaseRenderParams(),
		UniqueId:         deployer.env.UniqueId(""),
		Config:           *deployer.cfg,
	}
	//create deployment
	deployCfg, err := f.Open("chain-co/chain-co-statefulset.yaml")
	if err != nil {
		return err
	}

	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.statefulSets = append(deployer.statefulSets, statefulSet)

	deployer.pods, err = deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-chain-co-%s-pod", deployer.env.UniqueId("")))
	if err != nil {
		return err
	}

	//create service
	svcCfg, err := f.Open("chain-co/chain-co-headless.yaml")
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
