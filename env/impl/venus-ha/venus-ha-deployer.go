package venus_ha

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

	AuthUrl        string   `json:"-"`
	AdminToken     string   `json:"-"`
	BootstrapPeers []string `json:"-"`

	Replicas int `json:"replicas"`
}

type RenderParams struct {
	UniqueId string
	Config
}

func DefaultConfig() Config {
	return Config{
		Replicas: 1,
	}
}

var PluginInfo = types.PluginInfo{
	Name:        "venus-daemon-ha",
	Version:     version.Version(),
	Category:    types.Deploy,
	Repo:        "https://github.com/filecoin-project/venus.git",
	ImageTarget: "venus",
	Description: "",
}

var _ env.IVenusDeployer = (*VenusHADeployer)(nil)

type VenusHADeployer struct {
	outClusterEndpoint string
	endpoints          string
	env                *env.K8sEnvDeployer
	cfg                *Config

	svcEndpoint types.Endpoint

	pods        []corev1.Pod
	statefulSet []*appv1.StatefulSet
	svc         *corev1.Service
}

func NewVenusHADeployer(env *env.K8sEnvDeployer, replicas int, authUrl string, adminToken string, bootstrapPeers ...string) *VenusHADeployer {
	return &VenusHADeployer{
		env: env,
		cfg: &Config{
			Replicas:       replicas, //default
			AuthUrl:        authUrl,
			AdminToken:     adminToken,
			BootstrapPeers: bootstrapPeers,
		},
	}
}

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params Config) (env.IDeployer, error) {
	cfg, err := utils.MergeStructAndInterface(DefaultConfig(), cfg, params)
	if err != nil {
		return nil, err
	}
	return &VenusHADeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func (deployer *VenusHADeployer) Name() string {
	return PluginInfo.Name
}

func (deployer *VenusHADeployer) Deployment() []*appv1.Deployment {
	return nil
}

func (deployer *VenusHADeployer) Pods() []corev1.Pod {
	return deployer.pods
}

func (deployer *VenusHADeployer) StatefulSet() []*appv1.StatefulSet {
	return deployer.statefulSet
}

func (deployer *VenusHADeployer) Svc() *corev1.Service {
	return deployer.svc
}

func (deployer *VenusHADeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

//go:embed venus-node
var f embed.FS

func (deployer *VenusHADeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		UniqueId: deployer.env.UniqueId(""),
		Config:   *deployer.cfg,
	}

	//create statefulset
	deployCfg, err := f.Open("venus-node/venus-node-stateful-deployment.yaml")
	if err != nil {
		return err
	}
	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.statefulSet = append(deployer.statefulSet, statefulSet)

	deployer.pods, err = deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-%s-pod", deployer.env.UniqueId("")))
	if err != nil {
		return err
	}

	//create headless service
	svcCfg, err := f.Open("venus-node/venus-node-headless.yaml")
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
