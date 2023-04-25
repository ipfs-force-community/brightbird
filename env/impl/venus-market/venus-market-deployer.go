package venus_market

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

	NodeUrl     string `json:"-"`
	GatewayUrl  string `json:"-"`
	MessagerUrl string `json:"-"`
	AuthUrl     string `json:"-"`
	AuthToken   string `json:"-"`

	UseMysql bool `json:"useMysql"`
}

type RenderParams struct {
	env.BaseRenderParams
	Config

	UniqueId string
	MysqlDSN string
}

func DefaultConfig() Config {
	return Config{}
}

var PluginInfo = types.PluginInfo{
	Name:        string(env.VenusMarket),
	Version:     version.Version(),
	Category:    types.Deploy,
	Repo:        "https://github.com/filecoin-project/venus-market.git",
	ImageTarget: "venus-market",
	Description: "",
}

var _ env.IVenusMarketDeployer = (*VenusMarketDeployer)(nil)

type VenusMarketDeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	configMap    *corev1.ConfigMap
	pods         []corev1.Pod
	statefulSets []*appv1.StatefulSet
	svc          *corev1.Service
}

func NewVenusMarketDeployer(env *env.K8sEnvDeployer, authUrl, nodeUrl, gatewayUrl, messagerUrl, authToken string) *VenusMarketDeployer {
	return &VenusMarketDeployer{
		env: env,
		cfg: &Config{
			AuthUrl:     authUrl,
			NodeUrl:     nodeUrl,
			GatewayUrl:  gatewayUrl,
			MessagerUrl: messagerUrl,
			AuthToken:   authToken,
		},
	}
}

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params Config) (env.IDeployer, error) {
	cfg, err := utils.MergeStructAndInterface(DefaultConfig(), cfg, params)
	if err != nil {
		return nil, err
	}
	return &VenusMarketDeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func (deployer *VenusMarketDeployer) Name() string {
	return PluginInfo.Name
}

func (deployer *VenusMarketDeployer) Pods() []corev1.Pod {
	return deployer.pods
}

func (deployer *VenusMarketDeployer) Deployment() []*appv1.Deployment {
	return nil
}

func (deployer *VenusMarketDeployer) StatefulSets() []*appv1.StatefulSet {
	return deployer.statefulSets
}

func (deployer *VenusMarketDeployer) Svc() *corev1.Service {
	return deployer.svc
}

func (deployer *VenusMarketDeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

//go:embed venus-market
var f embed.FS

func (deployer *VenusMarketDeployer) Deploy(ctx context.Context) (err error) {
	renderParmas := RenderParams{
		BaseRenderParams: deployer.env.BaseRenderParams(),
		UniqueId:         deployer.env.UniqueId(""),
		Config:           *deployer.cfg,
	}
	if deployer.cfg.UseMysql {
		renderParmas.MysqlDSN = deployer.env.FormatMysqlConnection("venus-market-" + deployer.env.UniqueId(""))
	}
	//create configmap
	configMapCfg, err := f.Open("venus-market/venus-market-configmap.yaml")
	if err != nil {
		return err
	}
	deployer.configMap, err = deployer.env.CreateConfigMap(ctx, configMapCfg, renderParmas)
	if err != nil {
		return err
	}

	//create deployment
	deployCfg, err := f.Open("venus-market/venus-market-statefulset.yaml")
	if err != nil {
		return err
	}
	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParmas)
	if err != nil {
		return err
	}
	deployer.statefulSets = append(deployer.statefulSets, statefulSet)

	deployer.pods, err = deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-market-%s-pod", deployer.env.UniqueId("")))
	if err != nil {
		return err
	}

	//create service
	svcCfg, err := f.Open("venus-market/venus-market-headless.yaml")
	if err != nil {
		return err
	}
	deployer.svc, err = deployer.env.RunService(ctx, svcCfg, renderParmas)
	if err != nil {
		return err
	}

	deployer.svcEndpoint, err = deployer.env.WaitForServiceReady(ctx, deployer)
	if err != nil {
		return err
	}
	return nil
}
