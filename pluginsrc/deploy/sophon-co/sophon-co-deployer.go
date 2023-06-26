package sophonco

import (
	"context"
	"embed"
	"errors"
	"fmt"

	types2 "github.com/hunjixin/brightbird/types"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

////  The following types are used for components without configuration files or implemation with other lanaguage

// SophonCoConfig used to update sophon co
type SophonCoConfig struct { //nolint
	Nodes     []string
	AuthUrl   string
	AuthToken string
}

type Config struct {
	env.BaseConfig
	Replicas int `json:"replicas" description:"number of replicas"`

	AuthUrl    string   `json:"-"`
	AdminToken string   `json:"-"`
	Nodes      []string `json:"-"`
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

var PluginInfo = types2.PluginInfo{
	Name:        "sophon-co",
	Version:     version.Version(),
	PluginType:  types2.Deploy,
	Description: "",
	Repo:        "https://github.com/ipfs-force-community/sophon-co.git",
	ImageTarget: "sophon-co",
}

var _ env.IDeployer = (*SophonCoDeployer)(nil)

type SophonCoDeployer struct { //nolint
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types2.Endpoint

	statefulSetName string
	svcName         string
}

func NewSophonCoDeployer(env *env.K8sEnvDeployer, replicas int, authUrl string, ipEndpoints ...string) *SophonCoDeployer {
	return &SophonCoDeployer{
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
	return &SophonCoDeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func (deployer *SophonCoDeployer) InstanceName() (string, error) {
	return deployer.cfg.InstanceName, nil
}

func (deployer *SophonCoDeployer) Pods(ctx context.Context) ([]corev1.Pod, error) {
	return deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("sophon-co-%s-pod", env.UniqueId(deployer.env.TestID(), deployer.cfg.InstanceName)))
}

func (deployer *SophonCoDeployer) StatefulSet(ctx context.Context) (*appv1.StatefulSet, error) {
	return deployer.env.GetStatefulSet(ctx, deployer.statefulSetName)
}

func (deployer *SophonCoDeployer) Svc(ctx context.Context) (*corev1.Service, error) {
	return deployer.env.GetSvc(ctx, deployer.svcName)
}

func (deployer *SophonCoDeployer) SvcEndpoint() (types2.Endpoint, error) {
	return deployer.svcEndpoint, nil
}

func (deployer *SophonCoDeployer) Param(key string) (env.Params, error) {
	return env.Params{}, errors.New("no params")
}

//go:embed  sophon-co
var f embed.FS

func (deployer *SophonCoDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := deployer.buildRenderParams(deployer.cfg.Nodes, "")

	//create deployment
	deployCfg, err := f.Open("sophon-co/sophon-co-statefulset.yaml")
	if err != nil {
		return err
	}

	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.statefulSetName = statefulSet.GetName()

	//create service
	svcCfg, err := f.Open("sophon-co/sophon-co-headless.yaml")
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

func (deployer *SophonCoDeployer) GetConfig(ctx context.Context) (env.Params, error) {
	return env.ParamsFromVal(&SophonCoConfig{
		Nodes:     deployer.cfg.Nodes,
		AuthUrl:   deployer.cfg.AuthUrl,
		AuthToken: deployer.cfg.AdminToken,
	}), nil
}

func (deployer *SophonCoDeployer) Update(ctx context.Context, updateCfg interface{}) error {
	if updateCfg != nil {
		update := updateCfg.(*SophonCoConfig)
		//update params
		deployer.cfg.Nodes = update.Nodes
		deployer.cfg.AuthUrl = update.AuthUrl
		deployer.cfg.AdminToken = update.AuthToken

		//restart
		renderParams := deployer.buildRenderParams(update.Nodes, update.AuthUrl)
		// create deployment
		deployCfg, err := f.Open("sophon-co/sophon-co-statefulset.yaml")
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

func (deployer *SophonCoDeployer) buildRenderParams(nodes []string, authUrl string) RenderParams {
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
		UniqueId:        env.UniqueId(deployer.env.TestID(), deployer.cfg.InstanceName),
	}
}
