package marketclient

import (
	"context"
	"embed"
	"errors"
	"fmt"

	types2 "github.com/hunjixin/brightbird/types"

	"github.com/filecoin-project/go-address"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	"github.com/pelletier/go-toml/v2"
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
	Config

	NameSpace       string
	PrivateRegistry string
	Args            []string
	UniqueId        string
}

func DefaultConfig() Config {
	return Config{}
}

var PluginInfo = types2.PluginInfo{
	Name:        "market-client",
	Version:     version.Version(),
	PluginType:  types2.Deploy,
	Description: "",
	Repo:        "https://github.com/filecoin-project/venus-market.git",
	ImageTarget: "market-client",
}

type IMarketClientDeployer env.IDeployer

var _ IMarketClientDeployer = (*MarketClientDeployer)(nil)

type MarketClientDeployer struct { //nolint
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types2.Endpoint

	configMapName   string
	statefulSetName string
	svcName         string
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

func (deployer *MarketClientDeployer) InstanceName() (string, error) {
	return deployer.cfg.InstanceName, nil
}

func (deployer *MarketClientDeployer) Pods(ctx context.Context) ([]corev1.Pod, error) {
	return deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("market-client-%s-pod", env.UniqueId(deployer.env.TestID(), deployer.cfg.InstanceName)))
}

func (deployer *MarketClientDeployer) StatefulSet(ctx context.Context) (*appv1.StatefulSet, error) {
	return deployer.env.GetStatefulSet(ctx, deployer.statefulSetName)
}

func (deployer *MarketClientDeployer) Svc(ctx context.Context) (*corev1.Service, error) {
	return deployer.env.GetSvc(ctx, deployer.svcName)
}

func (deployer *MarketClientDeployer) SvcEndpoint() (types2.Endpoint, error) {
	return deployer.svcEndpoint, nil
}

func (deployer *MarketClientDeployer) Param(key string) (env.Params, error) {
	return env.Params{}, errors.New("no params")
}

//go:embed  market-client
var f embed.FS

func (deployer *MarketClientDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		NameSpace:       deployer.env.NameSpace(),
		PrivateRegistry: deployer.env.PrivateRegistry(),
		Args:            nil,
		UniqueId:        env.UniqueId(deployer.env.TestID(), deployer.cfg.InstanceName),
		Config:          *deployer.cfg,
	}
	//create configmap
	configMapCfg, err := f.Open("market-client/market-client-configmap.yaml")
	if err != nil {
		return err
	}
	configMap, err := deployer.env.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return err
	}

	deployer.configMapName = configMap.GetName()

	//create deployment
	deployCfg, err := f.Open("market-client/market-client-statefulset.yaml")
	if err != nil {
		return err
	}
	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.statefulSetName = statefulSet.GetName()

	//create service
	svcCfg, err := f.Open("market-client/market-client-headless.yaml")
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

func (deployer *MarketClientDeployer) GetConfig(ctx context.Context) (env.Params, error) {
	cfgData, err := deployer.env.GetConfigMap(ctx, deployer.configMapName, "config.toml")
	if err != nil {
		return env.Params{}, err
	}

	return env.ParamsFromVal(cfgData), nil
}

func (deployer *MarketClientDeployer) Update(ctx context.Context, updateCfg interface{}) error {
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
			_, err = deployer.env.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.marketclient/config.toml")
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
