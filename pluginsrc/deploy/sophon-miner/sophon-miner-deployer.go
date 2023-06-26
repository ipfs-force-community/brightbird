package sophonminer

import (
	"context"
	"embed"
	"errors"
	"fmt"

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

	NodeUrl    string `json:"-"`
	AuthUrl    string `json:"-"`
	GatewayUrl string `json:"-"`
	AuthToken  string `json:"-"`

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

var PluginInfo = types.PluginInfo{
	Name:        "sophon-miner",
	Version:     version.Version(),
	PluginType:  types.Deploy,
	Repo:        "https://github.com/ipfs-force-community/sophon-miner.git",
	ImageTarget: "sophon-miner",
	Description: "",
}

var _ env.IDeployer = (*SophonMinerDeployer)(nil)

type SophonMinerDeployer struct { //nolint
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	configMapName   string
	statefulSetName string
	svcName         string
}

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params Config) (env.IDeployer, error) {
	cfg, err := utils.MergeStructAndInterface(DefaultConfig(), cfg, params)
	if err != nil {
		return nil, err
	}
	return &SophonMinerDeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func (deployer *SophonMinerDeployer) InstanceName() (string, error) {
	return deployer.cfg.InstanceName, nil
}

func (deployer *SophonMinerDeployer) Pods(ctx context.Context) ([]corev1.Pod, error) {
	return deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("sophon-miner-%s-pod", env.UniqueId(deployer.env.TestID(), deployer.cfg.InstanceName)))
}

func (deployer *SophonMinerDeployer) StatefulSet(ctx context.Context) (*appv1.StatefulSet, error) {
	return deployer.env.GetStatefulSet(ctx, deployer.statefulSetName)
}

func (deployer *SophonMinerDeployer) Svc(ctx context.Context) (*corev1.Service, error) {
	return deployer.env.GetSvc(ctx, deployer.svcName)
}

func (deployer *SophonMinerDeployer) SvcEndpoint() (types.Endpoint, error) {
	return deployer.svcEndpoint, nil
}

func (deployer *SophonMinerDeployer) Param(key string) (env.Params, error) {
	return env.Params{}, errors.New("no params")
}

//go:embed sophon-miner
var f embed.FS

func (deployer *SophonMinerDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		NameSpace:       deployer.env.NameSpace(),
		PrivateRegistry: deployer.env.PrivateRegistry(),
		UniqueId:        env.UniqueId(deployer.env.TestID(), deployer.cfg.InstanceName),
		Config:          *deployer.cfg,
	}
	if deployer.cfg.UseMysql {
		renderParams.MysqlDSN = deployer.env.FormatMysqlConnection("sophon-miner-" + renderParams.UniqueId)
	}
	if len(renderParams.MysqlDSN) > 0 {
		err = deployer.env.ResourceMgr().EnsureDatabase(renderParams.MysqlDSN)
		if err != nil {
			return err
		}
	}
	//create configmap
	configMapCfg, err := f.Open("sophon-miner/sophon-miner-configmap.yaml")
	if err != nil {
		return err
	}
	cfgMap, err := deployer.env.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.configMapName = cfgMap.GetName()

	//create deployment
	deployCfg, err := f.Open("sophon-miner/sophon-miner-statefulset.yaml")
	if err != nil {
		return err
	}
	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.statefulSetName = statefulSet.GetName()

	//create service
	svcCfg, err := f.Open("sophon-miner/sophon-miner-headless.yaml")
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

func (deployer *SophonMinerDeployer) GetConfig(ctx context.Context) (env.Params, error) {
	cfgData, err := deployer.env.GetConfigMap(ctx, deployer.configMapName, "config.toml")
	if err != nil {
		return env.Params{}, err
	}

	return env.ParamsFromVal(cfgData), nil
}

func (deployer *SophonMinerDeployer) Update(ctx context.Context, updateCfg interface{}) error {
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
			_, err = deployer.env.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.sophon-miner/config.toml")
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
