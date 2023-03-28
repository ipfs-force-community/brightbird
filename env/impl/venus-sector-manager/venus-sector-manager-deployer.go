package venus_sector_manager

import (
	"context"
	"embed"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	env.BaseConfig

	NodeUrl     string `json:"-"`
	MessagerUrl string `json:"-"`
	MarketUrl   string `json:"-"`
	GatewayUrl  string `json:"-"`
	AuthUrl     string `json:"-"`
	AuthToken   string `json:"-"`

	DbPlugin     string `json:"dbPlugin"`
	DbDns        string `json:"dbDns"`
	SendAddress  string `json:"sendAddress"`
	MinerAddress int64  `json:"minerAddress"`
}

type RenderParams struct {
	TestID string
	Config
}

func DefaultConfig() Config {
	return Config{}
}

var PluginInfo = types.PluginInfo{
	Name:        string(env.VenusSectorManager),
	Version:     version.Version(),
	Category:    types.Deploy,
	Repo:        "https://github.com/ipfs-force-community/venus-cluster.git",
	ImageTarget: "venus-sector-manager",
	Description: "",
}

var _ env.IVenusSectorManagerDeployer = (*VenusSectorManagerDeployer)(nil)

type VenusSectorManagerDeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	configMap  *corev1.ConfigMap
	pods       []corev1.Pod
	deployment []*appv1.Deployment
	svc        *corev1.Service
}

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params Config) (env.IDeployer, error) {
	cfg, err := utils.MergeStructAndInterface(DefaultConfig(), cfg, params)
	if err != nil {
		return nil, err
	}
	return &VenusSectorManagerDeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func NewVenusSectorManagerDeployer(env *env.K8sEnvDeployer, nodeUrl, messagerUrl, marketUrl, gatewayUrl, authUrl,
	authToken, dbPlugin, sendAddress string, minerAddress int64) *VenusSectorManagerDeployer {
	dbDns := ""
	if dbPlugin == "sqlxdb" {
		dbDns = "root:123456@tcp(192.168.200.103:3306)/venus-sector-manager-" + env.TestID() + "?charset=utf8&parseTime=True&loc=Local"
	}
	return &VenusSectorManagerDeployer{
		env: env,
		cfg: &Config{
			NodeUrl:      nodeUrl,
			MessagerUrl:  messagerUrl,
			MarketUrl:    marketUrl,
			GatewayUrl:   gatewayUrl,
			AuthUrl:      authUrl,
			AuthToken:    authToken,
			DbPlugin:     dbPlugin,
			DbDns:        dbDns,
			SendAddress:  sendAddress,
			MinerAddress: minerAddress,
		},
	}
}

func (deployer *VenusSectorManagerDeployer) Name() string {
	return PluginInfo.Name
}

func (deployer *VenusSectorManagerDeployer) Pods() []corev1.Pod {
	return deployer.pods
}

func (deployer *VenusSectorManagerDeployer) Deployment() []*appv1.Deployment {
	return deployer.deployment
}

func (deployer *VenusSectorManagerDeployer) Svc() *corev1.Service {
	return deployer.svc
}

func (deployer *VenusSectorManagerDeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

var f embed.FS

func (deployer *VenusSectorManagerDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		TestID: deployer.env.TestID(),
		Config: *deployer.cfg,
	}

	// create configMap
	configMap, err := f.Open("venus-sector-manager/venus-sector-manager-configmap.yaml")
	if err != nil {
		return err
	}
	deployer.configMap, err = deployer.env.CreateConfigMap(ctx, configMap, renderParams)
	if err != nil {
		return err
	}

	// create deployment
	deployCfg, err := f.Open("venus-sector-manager/venus-sector-manager-deployment.yaml")
	if err != nil {
		return err
	}
	deployment, err := deployer.env.RunDeployment(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.deployment = append(deployer.deployment, deployment)

	// create service
	svcCfg, err := f.Open("venus-sector-manager/venus-sector-manager-service.yaml")
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
