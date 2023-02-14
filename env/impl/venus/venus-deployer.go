package venus

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	Replicas       int
	AuthUrl        string
	AdminToken     string
	BootstrapPeers []string
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
	Name:        "venus-daemon-simple",
	Version:     version.Version(),
	Description: "",
}

var _ env.IVenusDeployer = (*VenusDeployer)(nil)

type VenusDeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	pods       []corev1.Pod
	deployment []*appv1.Deployment
	svc        *corev1.Service
}

func NewVenusDeployer(env *env.K8sEnvDeployer, authUrl string, adminToken string, bootstrapPeers ...string) *VenusDeployer {
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

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params json.RawMessage) (env.IDeployer, error) {
	cfg, err := utils.MergeStructAndJson(DefaultConfig(), cfg, params)
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
	return deployer.deployment
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
		UniqueId: deployer.env.UniqueId(""),
		Config:   *deployer.cfg,
	}
	//create deployment
	deployCfg, err := f.Open("venus-node/venus-node-deployment.yaml")
	if err != nil {
		return err
	}
	deployment, err := deployer.env.RunDeployment(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.deployment = append(deployer.deployment, deployment)

	deployer.pods, err = deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-%s-pod", deployer.env.UniqueId("")))
	if err != nil {
		return err
	}

	//create service
	svcCfg, err := f.Open("venus-node/venus-node-service.yaml")
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
