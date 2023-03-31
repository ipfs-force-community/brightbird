package market_client

import (
	"context"
	"embed"
	"fmt"

	"github.com/filecoin-project/go-address"
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
	NodeToken   string `json:"-"`
	WalletUrl   string `json:"-"`
	WalletToken string `json:"-"`

	ClientAddr string `json:"clientAddr"`
}

type RenderParams struct {
	env.BaseRenderParams
	UniqueId string
	Config
}

func DefaultConfig() Config {
	return Config{}
}

var PluginInfo = types.PluginInfo{
	Name:        string(env.MarketClient),
	Version:     version.Version(),
	Category:    types.Deploy,
	Description: "",
	Repo:        "https://github.com/filecoin-project/venus-market.git",
	ImageTarget: "market-client",
}

type IMarketClientDeployer env.IDeployer

var _ IMarketClientDeployer = (*MarketClientDeployer)(nil)

type MarketClientDeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	configMap  *corev1.ConfigMap
	pods       []corev1.Pod
	deployment []*appv1.Deployment
	svc        *corev1.Service
}

func NewMarketClientDeployer(env *env.K8sEnvDeployer, nodeUrl, nodeToken, walletUrl, walletToken string, clientAddr address.Address) *MarketClientDeployer {
	return &MarketClientDeployer{
		env: env,
		cfg: &Config{
			NodeToken:   nodeToken,
			NodeUrl:     nodeUrl,
			WalletUrl:   walletUrl,
			WalletToken: walletToken,
			ClientAddr:  clientAddr.String(),
		},
	}
}

func DeployerFromConfig(env *env.K8sEnvDeployer, depCfg Config, frontCfg Config) (env.IDeployer, error) {
	cfg, err := utils.MergeStructAndInterface(DefaultConfig(), depCfg, frontCfg)
	if err != nil {
		return nil, err
	}
	return &MarketClientDeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func (deployer *MarketClientDeployer) Name() string {
	return PluginInfo.Name
}

func (deployer *MarketClientDeployer) Pods() []corev1.Pod {
	return deployer.pods
}

func (deployer *MarketClientDeployer) Deployment() []*appv1.Deployment {
	return deployer.deployment
}

func (deployer *MarketClientDeployer) Svc() *corev1.Service {
	return deployer.svc
}

func (deployer *MarketClientDeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

//go:embed  market-client
var f embed.FS

func (deployer *MarketClientDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		BaseRenderParams: deployer.env.BaseRenderParams(),
		UniqueId:         deployer.env.UniqueId(deployer.cfg.SvcMap[types.OutLabel]),
		Config:           *deployer.cfg,
	}
	//create configmap
	configMapCfg, err := f.Open("market-client/market-client-configmap.yaml")
	if err != nil {
		return err
	}
	deployer.configMap, err = deployer.env.CreateConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return err
	}

	//create deployment
	deployCfg, err := f.Open("market-client/market-client-deployment.yaml")
	if err != nil {
		return err
	}
	deployment, err := deployer.env.RunDeployment(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.deployment = append(deployer.deployment, deployment)

	deployer.pods, err = deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-market-client-%s-pod", deployer.env.UniqueId(deployer.cfg.SvcMap[types.OutLabel])))
	if err != nil {
		return err
	}

	//create service
	svcCfg, err := f.Open("market-client/market-client-service.yaml")
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
