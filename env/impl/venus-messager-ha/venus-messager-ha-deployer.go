package venus_messager_ha

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
	MysqlDSN   string `json:"-"`

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
	Name:        "venus-message-ha",
	Version:     version.Version(),
	Category:    types.Deploy,
	Repo:        "https://github.com/filecoin-project/venus-messager.git",
	ImageTarget: "venus-messager",
	Description: "",
}

var _ env.IVenusMessageDeployer = (*VenusMessagerHADeployer)(nil)

type VenusMessagerHADeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	pods       []corev1.Pod
	deployment []*appv1.Deployment
	svc        *corev1.Service
}

func NewVenusMessagerHADeployer(env *env.K8sEnvDeployer, replicas int, nodeUrl, authUrl, gatewayUrl, authToken string) *VenusMessagerHADeployer {
	return &VenusMessagerHADeployer{
		env: env,
		cfg: &Config{
			Replicas:   replicas, //default
			AuthUrl:    authUrl,
			GatewayUrl: gatewayUrl,
			AuthToken:  authToken,
			NodeUrl:    nodeUrl,
			MysqlDSN:   env.FormatMysqlConnection("venus-messager-ha-" + env.UniqueId("")),
		},
	}
}

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params Config) (env.IDeployer, error) {
	defaultCfg := DefaultConfig()
	defaultCfg.MysqlDSN = env.FormatMysqlConnection("venus-messager-ha-" + env.UniqueId(""))
	cfg, err := utils.MergeStructAndInterface(defaultCfg, cfg, params)
	if err != nil {
		return nil, err
	}
	return &VenusMessagerHADeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func (deployer *VenusMessagerHADeployer) Name() string {
	return PluginInfo.Name
}

func (deployer *VenusMessagerHADeployer) Pods() []corev1.Pod {
	return deployer.pods
}

func (deployer *VenusMessagerHADeployer) Deployment() []*appv1.Deployment {
	return deployer.deployment
}

func (deployer *VenusMessagerHADeployer) Svc() *corev1.Service {
	return deployer.svc
}

func (deployer *VenusMessagerHADeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

//go:embed venus-messager
var f embed.FS

func (deployer *VenusMessagerHADeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		BaseRenderParams: deployer.env.BaseRenderParams(),
		UniqueId:         deployer.env.UniqueId(""),
		Config:           *deployer.cfg,
	}

	//create database
	err = deployer.env.CreateDatabase(renderParams.MysqlDSN)
	if err != nil {
		return err
	}

	//create deploymnet for one node to push
	pushCfg, err := f.Open("venus-messager/venus-messager-push-deployment.yaml")
	if err != nil {
		return err
	}

	cfgCopy := renderParams
	cfgCopy.Replicas = 1
	deployment, err := deployer.env.RunDeployment(ctx, pushCfg, cfgCopy)
	if err != nil {
		return err
	}
	deployer.deployment = []*appv1.Deployment{deployment}
	if renderParams.Replicas > 1 {
		//deploy other node just service for others
		receiveCfg, err := f.Open("venus-messager/venus-messager-receive-deployment.yaml")
		if err != nil {
			return err
		}

		cfgCopy = renderParams
		cfgCopy.Replicas--
		deployment, err := deployer.env.RunDeployment(ctx, receiveCfg, cfgCopy)
		if err != nil {
			return err
		}
		deployer.deployment = append(deployer.deployment, deployment)
	}

	pods, err := deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-messager-%s-pod", deployer.env.UniqueId("")))
	if err != nil {
		return err
	}
	deployer.pods = pods

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
