package dropletmarket

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

	UseMysql bool `json:"useMysql" description:"true or false"`
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
	Name:        "droplet-market",
	Version:     version.Version(),
	PluginType:  types2.Deploy,
	Repo:        "https://github.com/ipfs-force-community/droplet.git",
	ImageTarget: "droplet-market",
	Description: "",
}

var _ env.IDeployer = (*DropletMarketDeployer)(nil)

type DropletMarketDeployer struct { //nolint
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types2.Endpoint

	configMapName   string
	statefulSetName string
	svcName         string
}

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params Config) (env.IDeployer, error) {
	cfg, err := utils.MergeStructAndInterface(DefaultConfig(), cfg, params)
	if err != nil {
		return nil, err
	}
	return &DropletMarketDeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func (deployer *DropletMarketDeployer) InstanceName() (string, error) {
	return PluginInfo.Name, nil
}

func (deployer *DropletMarketDeployer) Pods(ctx context.Context) ([]corev1.Pod, error) {
	return deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("droplet-market-%s-pod", env.UniqueId(deployer.env.TestID(), deployer.cfg.InstanceName)))
}

func (deployer *DropletMarketDeployer) StatefulSet(ctx context.Context) (*appv1.StatefulSet, error) {
	return deployer.env.GetStatefulSet(ctx, deployer.statefulSetName)
}

func (deployer *DropletMarketDeployer) Svc(ctx context.Context) (*corev1.Service, error) {
	return deployer.env.GetSvc(ctx, deployer.svcName)
}

func (deployer *DropletMarketDeployer) SvcEndpoint() (types2.Endpoint, error) {
	return deployer.svcEndpoint, nil
}

func (deployer *DropletMarketDeployer) Param(key string) (env.Params, error) {
	return env.Params{}, errors.New("no params")
}

//go:embed droplet-market
var f embed.FS

func (deployer *DropletMarketDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		NameSpace:       deployer.env.NameSpace(),
		PrivateRegistry: deployer.env.PrivateRegistry(),
		UniqueId:        env.UniqueId(deployer.env.TestID(), deployer.cfg.InstanceName),
		Config:          *deployer.cfg,
	}
	if deployer.cfg.UseMysql {
		renderParams.MysqlDSN = deployer.env.FormatMysqlConnection("droplet-market-" + renderParams.UniqueId)
	}
	//create configmap
	configMapCfg, err := f.Open("droplet-market/droplet-market-configmap.yaml")
	if err != nil {
		return err
	}
	configMap, err := deployer.env.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.configMapName = configMap.GetName()

	//create deployment
	deployCfg, err := f.Open("droplet-market/droplet-market-statefulset.yaml")
	if err != nil {
		return err
	}
	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.statefulSetName = statefulSet.GetName()

	//create service
	svcCfg, err := f.Open("droplet-market/droplet-market-headless.yaml")
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

func (deployer *DropletMarketDeployer) GetConfig(ctx context.Context) (env.Params, error) {
	cfgData, err := deployer.env.GetConfigMap(ctx, deployer.configMapName, "config.toml")
	if err != nil {
		return env.Params{}, err
	}

	return env.ParamsFromVal(cfgData), nil
}

func (deployer *DropletMarketDeployer) Update(ctx context.Context, updateCfg interface{}) error {
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
			_, err = deployer.env.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.droplet-market/config.toml")
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
