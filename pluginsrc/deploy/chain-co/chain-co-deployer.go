package chain_co

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

	AuthUrl    string   `json:"-"`
	AdminToken string   `json:"-"`
	Nodes      []string `json:"-"`

	Replicas int `json:"replicas"`
}

type RenderParams struct {
	Config

	NameSpace       string
	PrivateRegistry string
	Args            []string

	UniqueId string
}

func DefaultConfig() Config {
	return Config{Replicas: 0}
}

var PluginInfo = types.PluginInfo{
	Name:        "chain-co",
	Version:     version.Version(),
	Category:    types.Deploy,
	Description: "",
	Repo:        "https://github.com/ipfs-force-community/chain-co.git",
	ImageTarget: "chain-co",
}

var _ env.IDeployer = (*ChainCoDeployer)(nil)

type ChainCoDeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	statefulSetName string
	svcName         string
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

func (deployer *ChainCoDeployer) Pods(ctx context.Context) ([]corev1.Pod, error) {
	return deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-chain-co-%s-pod", deployer.env.UniqueId("")))
}

func (deployer *ChainCoDeployer) StatefulSet(ctx context.Context) (*appv1.StatefulSet, error) {
	return deployer.env.GetStatefulSet(ctx, deployer.statefulSetName)
}

func (deployer *ChainCoDeployer) Svc(ctx context.Context) (*corev1.Service, error) {
	return deployer.env.GetSvc(ctx, deployer.svcName)
}

func (deployer *ChainCoDeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

func (deployer *ChainCoDeployer) Param(key string) (interface{}, error) {
	return nil, errors.New("no params")
}

//go:embed  chain-co
var f embed.FS

func (deployer *ChainCoDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := deployer.buildRenderParams(deployer.cfg.Nodes, "")

	//create deployment
	deployCfg, err := f.Open("chain-co/chain-co-statefulset.yaml")
	if err != nil {
		return err
	}

	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.statefulSetName = statefulSet.GetName()

	//create service
	svcCfg, err := f.Open("chain-co/chain-co-headless.yaml")
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

func (deployer *ChainCoDeployer) GetConfig(ctx context.Context) (interface{}, error) {
	return &env.ChainCoConfig{
		Nodes:     deployer.cfg.Nodes,
		AuthUrl:   deployer.cfg.AuthUrl,
		AuthToken: deployer.cfg.AdminToken,
	}, nil
}

func (deployer *ChainCoDeployer) Update(ctx context.Context, updateCfg interface{}) error {
	if updateCfg != nil {
		update := updateCfg.(*env.ChainCoConfig)
		//update params
		deployer.cfg.Nodes = update.Nodes
		deployer.cfg.AuthUrl = update.AuthUrl
		deployer.cfg.AdminToken = update.AuthToken

		//restart
		renderParams := deployer.buildRenderParams(update.Nodes, update.AuthUrl)
		// create deployment
		deployCfg, err := f.Open("chain-co/chain-co-statefulset.yaml")
		if err != nil {
			return err
		}

		_, err = deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
		return err
	}

	err := deployer.env.UpdateStatefulSets(ctx, deployer.statefulSetName)
	if err != nil {
		return err
	}
	return nil
}

func (deployer *ChainCoDeployer) buildRenderParams(nodes []string, authUrl string) RenderParams {
	var args []string
	for _, node := range deployer.cfg.Nodes {
		args = append(args, "--node")
		args = append(args, node)
	}

	if len(authUrl) > 0 {
		args = append(args, "--auth")
		args = append(args, deployer.cfg.AuthUrl)
	} else {
		if len(deployer.cfg.AuthUrl) > 0 {
			args = append(args, "--auth")
			args = append(args, deployer.cfg.AuthUrl)
		}
	}

	return RenderParams{
		NameSpace:       deployer.env.NameSpace(),
		Config:          *deployer.cfg,
		PrivateRegistry: deployer.env.PrivateRegistry(),
		Args:            args,
		UniqueId:        deployer.env.UniqueId(""),
	}
}
