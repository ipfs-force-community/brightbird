package venus_wallet

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
	GatewayUrl      string
	AuthToken       string
	SupportAccounts []string
	//use for annotate service name
	SvcMap map[string]string
}

type RenderParams struct {
	UniqueId string
	Config
}

func DefaultConfig() Config {
	return Config{}
}

var PluginInfo = types.PluginInfo{
	Name:        string(env.VenusWallet),
	Version:     version.Version(),
	Description: "",
}

var _ env.IVenusWalletDeployer = (*VenusWalletDeployer)(nil)

type VenusWalletDeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	configMap  *corev1.ConfigMap
	pods       []corev1.Pod
	deployment []*appv1.Deployment
	svc        *corev1.Service
}

func NewVenusWalletDeployer(env *env.K8sEnvDeployer, gatewayUrl, authToken string, supportAccounts ...string) *VenusWalletDeployer {
	return &VenusWalletDeployer{
		env: env,
		cfg: &Config{
			GatewayUrl:      gatewayUrl,
			AuthToken:       authToken,
			SupportAccounts: supportAccounts,
		},
	}
}

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params json.RawMessage) (env.IDeployer, error) {
	cfg, err := utils.MergeStructAndJson(DefaultConfig(), cfg, params)
	if err != nil {
		return nil, err
	}
	return &VenusWalletDeployer{
		env: env,
		cfg: &cfg,
	}, nil
}
func (deployer *VenusWalletDeployer) Name() string {
	return PluginInfo.Name
}

func (deployer *VenusWalletDeployer) Pods() []corev1.Pod {
	return deployer.pods
}

func (deployer *VenusWalletDeployer) Deployment() []*appv1.Deployment {
	return deployer.deployment
}

func (deployer *VenusWalletDeployer) Svc() *corev1.Service {
	return deployer.svc
}

func (deployer *VenusWalletDeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

//go:embed venus-wallet
var f embed.FS

func (deployer *VenusWalletDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		UniqueId: deployer.env.UniqueId(deployer.cfg.SvcMap[types.OutLabel]),
		Config:   *deployer.cfg,
	}
	//create configmap
	configMapCfg, err := f.Open("venus-wallet/venus-wallet-configmap.yaml")
	if err != nil {
		return err
	}
	deployer.configMap, err = deployer.env.CreateConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return err
	}

	//create deployment
	deployCfg, err := f.Open("venus-wallet/venus-wallet-deployment.yaml")
	if err != nil {
		return err
	}
	deployment, err := deployer.env.RunDeployment(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.deployment = append(deployer.deployment, deployment)

	deployer.pods, err = deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-wallet-%s-pod", deployer.env.UniqueId(deployer.cfg.SvcMap[types.OutLabel])))
	if err != nil {
		return err
	}

	//create service
	svcCfg, err := f.Open("venus-wallet/venus-wallet-service.yaml")
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
