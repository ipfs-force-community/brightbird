package venus_market

import (
	"context"
	"embed"
	"errors"
	"fmt"

	types2 "github.com/hunjixin/brightbird/types"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	"github.com/pelletier/go-toml"
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
	Config

	NameSpace       string
	PrivateRegistry string
	Args            []string

	UniqueId string
	MysqlDSN string
}

func DefaultConfig() Config {
	return Config{}
}

var PluginInfo = types2.PluginInfo{
	Name:        "venus-market",
	Version:     version.Version(),
	PluginType:  types2.Deploy,
	Repo:        "https://github.com/filecoin-project/venus-market.git",
	ImageTarget: "venus-market",
	Description: "",
}

var _ env.IDeployer = (*VenusMarketDeployer)(nil)

type VenusMarketDeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types2.Endpoint

	configMapName   string
	statefulSetName string
	svcName         string
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

func (deployer *VenusMarketDeployer) InstanceName() (string, error) {
	return PluginInfo.Name, nil
}

func (deployer *VenusMarketDeployer) Pods(ctx context.Context) ([]corev1.Pod, error) {
	return deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-market-%s-pod", env.UniqueId(deployer.env.TestID(), deployer.cfg.InstanceName)))
}

func (deployer *VenusMarketDeployer) StatefulSet(ctx context.Context) (*appv1.StatefulSet, error) {
	return deployer.env.GetStatefulSet(ctx, deployer.statefulSetName)
}

func (deployer *VenusMarketDeployer) Svc(ctx context.Context) (*corev1.Service, error) {
	return deployer.env.GetSvc(ctx, deployer.svcName)
}

func (deployer *VenusMarketDeployer) SvcEndpoint() (types2.Endpoint, error) {
	return deployer.svcEndpoint, nil
}

func (deployer *VenusMarketDeployer) Param(key string) (env.Params, error) {
	return env.Params{}, errors.New("no params")
}

//go:embed venus-market
var f embed.FS

func (deployer *VenusMarketDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		NameSpace:       deployer.env.NameSpace(),
		PrivateRegistry: deployer.env.PrivateRegistry(),
		UniqueId:        env.UniqueId(deployer.env.TestID(), deployer.cfg.InstanceName),
		Config:          *deployer.cfg,
	}
	if deployer.cfg.UseMysql {
		renderParams.MysqlDSN = deployer.env.FormatMysqlConnection("venus-market-" + renderParams.UniqueId)
	}
	//create configmap
	configMapCfg, err := f.Open("venus-market/venus-market-configmap.yaml")
	if err != nil {
		return err
	}
	configMap, err := deployer.env.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.configMapName = configMap.GetName()

	//create deployment
	deployCfg, err := f.Open("venus-market/venus-market-statefulset.yaml")
	if err != nil {
		return err
	}
	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.statefulSetName = statefulSet.GetName()

	//create service
	svcCfg, err := f.Open("venus-market/venus-market-headless.yaml")
	if err != nil {
		return err
	}
	svc, err := deployer.env.RunService(ctx, svcCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.svcName = svc.Name

	deployer.svcEndpoint, err = deployer.env.WaitForServiceReady(ctx, deployer)
	if err != nil {
		return err
	}
	return nil
}

func (deployer *VenusMarketDeployer) GetConfig(ctx context.Context) (env.Params, error) {
	cfgData, err := deployer.env.GetConfigMap(ctx, deployer.configMapName, "config.toml")
	if err != nil {
		return env.Params{}, err
	}

	return env.ParamsFromVal(cfgData), nil
}

func (deployer *VenusMarketDeployer) Update(ctx context.Context, updateCfg interface{}) error {
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
			_, err = deployer.env.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.venusmarket/config.toml")
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
