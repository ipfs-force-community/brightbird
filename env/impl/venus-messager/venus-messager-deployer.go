package venus_messager

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

	NodeUrl    string `json:"-"`
	GatewayUrl string `json:"-"`
	AuthUrl    string `json:"-"`
	AuthToken  string `json:"-"`
}

type RenderParams struct {
	UniqueId string
	Config
}

func DefaultConfig() Config {
	return Config{}
}

var PluginInfo = types.PluginInfo{
	Name:        "venus-message-simple",
	Version:     version.Version(),
	Category:    types.Deploy,
	Repo:        "https://github.com/filecoin-project/venus-messager.git",
	Description: "",
}

var _ env.IVenusMessageDeployer = (*VenusMessagerDeployer)(nil)

type VenusMessagerDeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	pods       []corev1.Pod
	deployment []*appv1.Deployment
	svc        *corev1.Service
}

func NewVenusMessagerDeployer(env *env.K8sEnvDeployer, nodeUrl, authUrl, gatewayUrl, authToken string) *VenusMessagerDeployer {
	return &VenusMessagerDeployer{
		env: env,
		cfg: &Config{
			AuthUrl:    authUrl,
			AuthToken:  authToken,
			NodeUrl:    nodeUrl,
			GatewayUrl: gatewayUrl,
		},
	}
}

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params Config) (env.IDeployer, error) {
	cfg, err := utils.MergeStructAndInterface(DefaultConfig(), cfg, params)
	if err != nil {
		return nil, err
	}
	return &VenusMessagerDeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func (deployer *VenusMessagerDeployer) Name() string {
	return PluginInfo.Name
}

func (deployer *VenusMessagerDeployer) Pods() []corev1.Pod {
	return deployer.pods
}

func (deployer *VenusMessagerDeployer) Deployment() []*appv1.Deployment {
	return deployer.deployment
}

func (deployer *VenusMessagerDeployer) Svc() *corev1.Service {
	return deployer.svc
}

func (deployer *VenusMessagerDeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

//go:embed venus-messager
var f embed.FS

func (deployer *VenusMessagerDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		UniqueId: deployer.env.UniqueId(""),
		Config:   *deployer.cfg,
	}

	//create  deployment
	deployCfg, err := f.Open("venus-messager/venus-messager-sqlite-deployment.yaml")
	if err != nil {
		return err
	}
	deployment, err := deployer.env.RunDeployment(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.deployment = append(deployer.deployment, deployment)

	deployer.pods, err = deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-messager-%s-pod", deployer.env.UniqueId("")))
	if err != nil {
		return err
	}

	//create service
	svcCfg, err := f.Open("venus-messager/venus-messager-service.yaml")
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
