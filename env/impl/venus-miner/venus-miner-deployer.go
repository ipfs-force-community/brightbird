package venus_miner

import (
	"context"
	"embed"
	"fmt"

	mConfig "github.com/filecoin-project/venus-miner/node/config"
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

var PluginInfo = types.PluginInfo{
	Name:        string(env.VenusMiner),
	Version:     version.Version(),
	Category:    types.Deploy,
	Repo:        "https://github.com/filecoin-project/venus-miner.git",
	ImageTarget: "venus-miner",
	Description: "",
}

var _ env.IVenusMinerDeployer = (*VenusMinerDeployer)(nil)

type VenusMinerDeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	configMapName   string
	statefulSetName string
	svcName         string
}

func NewVenusMinerDeployer(env *env.K8sEnvDeployer, nodeUrl, authUrl, authToken string, useMysql bool) *VenusMinerDeployer {

	return &VenusMinerDeployer{
		env: env,
		cfg: &Config{
			NodeUrl:   nodeUrl,
			AuthToken: authToken,
			AuthUrl:   authUrl,
			UseMysql:  useMysql,
		},
	}
}

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params Config) (env.IVenusMinerDeployer, error) {
	cfg, err := utils.MergeStructAndInterface(DefaultConfig(), cfg, params)
	if err != nil {
		return nil, err
	}
	return &VenusMinerDeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func (deployer *VenusMinerDeployer) Name() string {
	return PluginInfo.Name
}

func (deployer *VenusMinerDeployer) Pods(ctx context.Context) ([]corev1.Pod, error) {
	return deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-miner-%s-pod", deployer.env.UniqueId("")))
}

func (deployer *VenusMinerDeployer) StatefulSet(ctx context.Context) (*appv1.StatefulSet, error) {
	return deployer.env.GetStatefulSet(ctx, deployer.statefulSetName)
}

func (deployer *VenusMinerDeployer) Svc(ctx context.Context) (*corev1.Service, error) {
	return deployer.env.GetSvc(ctx, deployer.svcName)
}

func (deployer *VenusMinerDeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

//go:embed venus-miner
var f embed.FS

func (deployer *VenusMinerDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		NameSpace:       deployer.env.NameSpace(),
		PrivateRegistry: deployer.env.PrivateRegistry(),
		UniqueId:        deployer.env.UniqueId(""),
		Config:          *deployer.cfg,
	}
	if deployer.cfg.UseMysql {
		renderParams.MysqlDSN = deployer.env.FormatMysqlConnection("venus-miner-" + deployer.env.UniqueId(""))
	}
	if len(renderParams.MysqlDSN) > 0 {
		deployer.env.ResourceMgr().EnsureDatabase(renderParams.MysqlDSN)
	}
	//create configmap
	configMapCfg, err := f.Open("venus-miner/venus-miner-configmap.yaml")
	if err != nil {
		return err
	}
	cfgMap, err := deployer.env.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.configMapName = cfgMap.GetName()

	//create deployment
	deployCfg, err := f.Open("venus-miner/venus-miner-statefulset.yaml")
	if err != nil {
		return err
	}
	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.statefulSetName = statefulSet.GetName()

	//create service
	svcCfg, err := f.Open("venus-miner/venus-miner-headless.yaml")
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

func (deployer *VenusMinerDeployer) GetConfig(ctx context.Context) (interface{}, error) {
	cfgData, err := deployer.env.GetConfigMap(ctx, deployer.configMapName, "config.toml")
	if err != nil {
		return nil, err
	}

	cfg := &mConfig.MinerConfig{}
	err = toml.Unmarshal(cfgData, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (deployer *VenusMinerDeployer) Update(ctx context.Context, updateCfg interface{}) error {
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
			_, err = deployer.env.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.venusminer/config.toml")
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
