package venus_gateway

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

	AuthUrl    string           `json:"-"`
	AdminToken types.AdminToken `json:"-"`

	Replicas int `json:"replicas"`
}

type RenderParams struct {
	env.BaseRenderParams
	Config

	UniqueId string
}

func DefaultConfig() Config {
	return Config{
		Replicas: 1,
	}
}

var PluginInfo = types.PluginInfo{
	Name:        string(env.VenusGateway),
	Version:     version.Version(),
	Category:    types.Deploy,
	Repo:        "https://github.com/ipfs-force-community/venus-gateway.git",
	ImageTarget: "venus-gateway",
	Description: "",
}

var _ env.IVenusGatewayDeployer = (*VenusGatewayDeployer)(nil)

type VenusGatewayDeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	pods         []corev1.Pod
	statefulSets []*appv1.StatefulSet
	svc          *corev1.Service
}

func NewVenusGatewayDeployer(env *env.K8sEnvDeployer, replicas int, authUrl string) *VenusGatewayDeployer {
	return &VenusGatewayDeployer{
		env: env,
		cfg: &Config{
			Replicas: replicas, //default
			AuthUrl:  authUrl,
		},
	}
}

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params Config) (env.IDeployer, error) {
	cfg, err := utils.MergeStructAndInterface(DefaultConfig(), cfg, params)
	if err != nil {
		return nil, err
	}
	return &VenusGatewayDeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func (deployer *VenusGatewayDeployer) Name() string {
	return PluginInfo.Name
}

func (deployer *VenusGatewayDeployer) Pods() []corev1.Pod {
	return deployer.pods
}

func (deployer *VenusGatewayDeployer) Deployment() []*appv1.Deployment {
	return nil
}

func (deployer *VenusGatewayDeployer) StatefulSets() []*appv1.StatefulSet {
	return deployer.statefulSets
}

func (deployer *VenusGatewayDeployer) Svc() *corev1.Service {
	return deployer.svc
}

func (deployer *VenusGatewayDeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

//go:embed venus-gateway
var f embed.FS

func (deployer *VenusGatewayDeployer) Deploy(ctx context.Context) error {
	renderParams := RenderParams{
		BaseRenderParams: deployer.env.BaseRenderParams(),
		UniqueId:         deployer.env.UniqueId(""),
		Config:           *deployer.cfg,
	}
	//create deployment
	deployCfg, err := f.Open("venus-gateway/venus-gateway-statefulset.yaml")
	if err != nil {
		return err
	}
	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.statefulSets = append(deployer.statefulSets, statefulSet)

	pods, err := deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-gateway-%s-pod", deployer.env.UniqueId("")))
	if err != nil {
		return err
	}
	deployer.pods = pods

	//create service
	svcCfg, err := f.Open("venus-gateway/venus-gateway-headless.yaml")
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
