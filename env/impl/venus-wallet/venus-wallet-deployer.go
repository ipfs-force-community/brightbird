package venus_wallet

import (
	"context"
	"embed"
	"fmt"

	wConfig "github.com/filecoin-project/venus-wallet/config"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	"github.com/pelletier/go-toml"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	env.BaseConfig

	GatewayUrl string `json:"-"`
	UserToken  string `json:"-"`

	UserName        string `json:"userName"`
	CreateIfNotExit bool   `json:"createIfNotExit"`
}

type RenderParams struct {
	Config

	NameSpace       string
	PrivateRegistry string
	Args            []string

	UniqueId string
}

func DefaultConfig() Config {
	return Config{}
}

var PluginInfo = types.PluginInfo{
	Name:        string(env.VenusWallet),
	Version:     version.Version(),
	Category:    types.Deploy,
	Repo:        "https://github.com/filecoin-project/venus-wallet.git",
	ImageTarget: "venus-wallet",
	Description: "",
}

var _ env.IDeployer = (*VenusWalletDeployer)(nil)

type VenusWalletDeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	configMapName   string
	statefulSetName string
	svcName         string
}

func NewVenusWalletDeployer(env *env.K8sEnvDeployer, gatewayUrl, userToken string, supportAccounts ...string) *VenusWalletDeployer {
	return &VenusWalletDeployer{
		env: env,
		cfg: &Config{
			GatewayUrl: gatewayUrl,
			UserToken:  userToken,
		},
	}
}

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params Config) (env.IDeployer, error) {
	cfg, err := utils.MergeStructAndInterface(DefaultConfig(), cfg, params)
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

func (deployer *VenusWalletDeployer) Pods(ctx context.Context) ([]corev1.Pod, error) {
	return deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-wallet-%s-pod", deployer.env.UniqueId(deployer.cfg.SvcMap[types.OutLabel])))
}

func (deployer *VenusWalletDeployer) StatefulSet(ctx context.Context) (*appv1.StatefulSet, error) {
	return deployer.env.GetStatefulSet(ctx, deployer.statefulSetName)
}

func (deployer *VenusWalletDeployer) Svc(ctx context.Context) (*corev1.Service, error) {
	return deployer.env.GetSvc(ctx, deployer.svcName)
}

func (deployer *VenusWalletDeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

//go:embed venus-wallet
var f embed.FS

func (deployer *VenusWalletDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		NameSpace:       deployer.env.NameSpace(),
		PrivateRegistry: deployer.env.PrivateRegistry(),
		UniqueId:        deployer.env.UniqueId(deployer.cfg.SvcMap[types.OutLabel]),
		Config:          *deployer.cfg,
	}
	//create configmap
	configMapCfg, err := f.Open("venus-wallet/venus-wallet-configmap.yaml")
	if err != nil {
		return err
	}
	configMap, err := deployer.env.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.configMapName = configMap.GetName()

	//create deployment
	deployCfg, err := f.Open("venus-wallet/venus-wallet-statefulset.yaml")
	if err != nil {
		return err
	}
	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.statefulSetName = statefulSet.GetName()

	//create service
	svcCfg, err := f.Open("venus-wallet/venus-wallet-headless.yaml")
	if err != nil {
		return err
	}
	svc, err := deployer.env.RunService(ctx, svcCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.svcName = svc.GetName()

	deployer.svcEndpoint, err = deployer.env.WaitForServiceReady(ctx, deployer)
	if err != nil {
		return err
	}
	return nil
}

func (deployer *VenusWalletDeployer) GetConfig(ctx context.Context) (interface{}, error) {
	cfgData, err := deployer.env.GetConfigMap(ctx, deployer.configMapName, "config.toml")
	if err != nil {
		return nil, err
	}

	cfg := &wConfig.Config{}
	err = toml.Unmarshal(cfgData, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (deployer *VenusWalletDeployer) Update(ctx context.Context, updateCfg interface{}) error {
	if updateCfg != nil {
		cfgData, err := toml.Marshal(updateCfg)
		if err != nil {
			return err
		}
		err = deployer.env.SetConfigMap(ctx, deployer.configMapName, "config.toml", cfgData)
		if err != nil {
			return err
		}

		pods, err := deployer.Pods(ctx)
		if err != nil {
			return nil
		}
		for _, pod := range pods {
			_, err = deployer.env.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.venus_wallet/config.toml")
			if err != nil {
				return err
			}
		}
	}

	err := deployer.env.UpdateStatefulSets(ctx, deployer.statefulSetName)
	if err != nil {
		return err
	}
	return nil
}
