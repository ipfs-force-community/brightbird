package venuswallet

import (
	"context"
	"embed"
	"fmt"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"github.com/pelletier/go-toml"
)

type Config struct {
	env.BaseConfig

	GatewayUrl string `json:"gatewayUrl"`
	UserToken  string `json:"userToekn"`
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
	Name:        "venus-wallet",
	Version:     version.Version(),
	PluginType:  types.Deploy,
	Repo:        "https://github.com/filecoin-project/venus-wallet.git",
	ImageTarget: "venus-wallet",
	Description: "",
}

type VenusWalletDeployParams struct {
	Cfg             *Config
	StatefulSetName string
	ConfigMapName   string
	SVCName         string
	SvcEndpoint     types.Endpoint
}

type VenusWalletDeployer struct { //nolint
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	configMapName   string
	statefulSetName string
	svcName         string
}

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config) (*VenusWalletDeployer, error) {
	return &VenusWalletDeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

//go:embed venus-wallet
var f embed.FS

func (deployer *VenusWalletDeployer) Deploy(ctx context.Context) (*VenusWalletDeployParams, error) {
	renderParams := RenderParams{
		NameSpace:       deployer.env.NameSpace(),
		PrivateRegistry: deployer.env.PrivateRegistry(),
		UniqueId:        env.UniqueId(deployer.env.TestID(), deployer.cfg.InstanceName),
		Config:          *deployer.cfg,
	}
	//create configmap
	configMapCfg, err := f.Open("venus-wallet/venus-wallet-configmap.yaml")
	if err != nil {
		return nil, err
	}
	configMap, err := deployer.env.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return nil, err
	}
	deployer.configMapName = configMap.GetName()

	//create deployment
	deployCfg, err := f.Open("venus-wallet/venus-wallet-statefulset.yaml")
	if err != nil {
		return nil, err
	}
	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return nil, err
	}
	deployer.statefulSetName = statefulSet.GetName()

	pods, err := deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-wallet-%s-pod", env.UniqueId(deployer.env.TestID(), deployer.cfg.InstanceName)))
	if err != nil {
		return nil, err
	}
	//create service
	svcCfg, err := f.Open("venus-wallet/venus-wallet-headless.yaml")
	if err != nil {
		return nil, err
	}
	svc, err := deployer.env.RunService(ctx, svcCfg, renderParams)
	if err != nil {
		return nil, err
	}
	deployer.svcName = svc.GetName()

	deployer.svcEndpoint, err = deployer.env.WaitForServiceReady(ctx, svc, pods)
	if err != nil {
		return nil, err
	}
	return &VenusWalletDeployParams{
		Cfg:             deployer.cfg,
		StatefulSetName: deployer.statefulSetName,
		ConfigMapName:   deployer.configMapName,
		SVCName:         deployer.svcName,
		SvcEndpoint:     deployer.svcEndpoint,
	}, nil
}

func (deployer *VenusWalletDeployer) GetConfig(ctx context.Context) (env.Params, error) {
	cfgData, err := deployer.env.GetConfigMap(ctx, deployer.configMapName, "config.toml")
	if err != nil {
		return env.Params{}, err
	}

	return env.ParamsFromVal(cfgData), nil
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
