package venus_auth

import (
	"context"
	"embed"
	_ "embed"
	"fmt"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Config struct{}

type RenderParams struct {
	UniqueId string
	Replicas int
}

func DefaultConfig() Config {
	return Config{}
}

var PluginInfo = types.PluginInfo{
	Name:        "venus-auth-simple",
	Version:     version.Version(),
	Category:    types.Deploy,
	Repo:        "https://github.com/filecoin-project/venus-auth.git",
	ImageTarget: "venus-auth",
	Description: "",
}

var _ env.IVenusAuthDeployer = (*VenusAuthDeployer)(nil)

type VenusAuthDeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	pods       []corev1.Pod
	deployment []*appv1.Deployment
	svc        *corev1.Service
}

func NewVenusAuthDeployer(env *env.K8sEnvDeployer) *VenusAuthDeployer {
	return &VenusAuthDeployer{
		env: env,
		cfg: &Config{},
	}
}

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params Config) (env.IDeployer, error) {
	cfg, err := utils.MergeStructAndInterface(DefaultConfig(), cfg, params)
	if err != nil {
		return nil, err
	}
	return &VenusAuthDeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func (deployer *VenusAuthDeployer) Name() string {
	return PluginInfo.Name
}

func (deployer *VenusAuthDeployer) Pods() []corev1.Pod {
	return deployer.pods
}

func (deployer *VenusAuthDeployer) Deployment() []*appv1.Deployment {
	return deployer.deployment
}

func (deployer *VenusAuthDeployer) Svc() *corev1.Service {
	return deployer.svc
}

func (deployer *VenusAuthDeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

//go:embed venus-auth
var f embed.FS

func (deployer *VenusAuthDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		UniqueId: deployer.env.UniqueId(""),
		Replicas: 1,
	}

	//create deployment
	deployCfg, err := f.Open("venus-auth/venus-auth-deployment.yaml")
	if err != nil {
		return err
	}

	deployment, err := deployer.env.RunDeployment(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.deployment = append(deployer.deployment, deployment)

	deployer.pods, err = deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-auth-%s-pod", deployer.env.UniqueId("")))
	if err != nil {
		return err
	}

	//create service
	svcCfg, err := f.Open("venus-auth/venus-auth-service.yaml")
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
