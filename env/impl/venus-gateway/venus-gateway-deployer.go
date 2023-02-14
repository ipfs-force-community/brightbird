package venus_gateway

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
	Replicas int
	AuthUrl  string
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
	Name:        string(env.VenusGateway),
	Version:     version.Version(),
	Description: "",
}

var _ env.IVenusGatewayDeployer = (*VenusGatewayDeployer)(nil)

type VenusGatewayDeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	pods       []corev1.Pod
	deployment []*appv1.Deployment
	svc        *corev1.Service
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

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params json.RawMessage) (env.IDeployer, error) {
	cfg, err := utils.MergeStructAndJson(DefaultConfig(), cfg, params)
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
	return deployer.deployment
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
		UniqueId: deployer.env.UniqueId(""),
		Config:   *deployer.cfg,
	}

	//create deployment
	deployCfg, err := f.Open("venus-gateway/venus-gateway-deployment.yaml")
	if err != nil {
		return err
	}
	deployment, err := deployer.env.RunDeployment(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.deployment = append(deployer.deployment, deployment)

	pods, err := deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-gateway-%s-pod", deployer.env.UniqueId("")))
	if err != nil {
		return err
	}
	deployer.pods = pods

	//create service
	svcCfg, err := f.Open("venus-gateway/venus-gateway-service.yaml")
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
