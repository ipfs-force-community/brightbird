package venus

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

	AuthUrl        string           `json:"-"`
	AdminToken     types.AdminToken `json:"-"`
	BootstrapPeers []string         `json:"-"`
	Replicas       int              `json:"replicas"`

	NetType string `json:"netType"`
}

type RenderParams struct {
	env.BaseRenderParams
	Config

	UniqueId string
}

func DefaultConfig() Config {
	return Config{
		Replicas: 1,
		NetType:  "force",
	}
}

var PluginInfo = types.PluginInfo{
	Name:        "venus-daemon-simple",
	Version:     version.Version(),
	Category:    types.Deploy,
	Repo:        "https://github.com/filecoin-project/venus.git",
	ImageTarget: "venus",
	Description: "",
}

var _ env.IVenusDeployer = (*VenusDeployer)(nil)

type VenusDeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	pods         []corev1.Pod
	statefulSets []*appv1.StatefulSet
	svc          *corev1.Service
}

func NewVenusDeployer(env *env.K8sEnvDeployer, authUrl string, adminToken types.AdminToken, bootstrapPeers ...string) *VenusDeployer {
	return &VenusDeployer{
		env: env,
		cfg: &Config{
			Replicas:       1, //default
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
	return &VenusDeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func (deployer *VenusDeployer) Name() string {
	return PluginInfo.Name
}

func (deployer *VenusDeployer) Pods() []corev1.Pod {
	return deployer.pods
}

func (deployer *VenusDeployer) Deployment() []*appv1.Deployment {
	return nil
}

func (deployer *VenusDeployer) StatefulSets() []*appv1.StatefulSet {
	return deployer.statefulSets
}

func (deployer *VenusDeployer) Svc() *corev1.Service {
	return deployer.svc
}

func (deployer *VenusDeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

//go:embed venus-node
var f embed.FS

func (deployer *VenusDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		BaseRenderParams: deployer.env.BaseRenderParams(),
		UniqueId:         deployer.env.UniqueId(""),
		Config:           *deployer.cfg,
	}
	//create deployment
	deployCfg, err := f.Open("venus-node/venus-node-statefulset.yaml")
	if err != nil {
		return err
	}
	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.statefulSets = append(deployer.statefulSets, statefulSet)

	deployer.pods, err = deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-%s-pod", deployer.env.UniqueId("")))
	if err != nil {
		return err
	}

	//create service
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
