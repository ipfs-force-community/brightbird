package venus_sector_manager

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	"github.com/ipfs-force-community/venus-cluster/venus-sector-manager/modules"
	"github.com/pelletier/go-toml"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	env.BaseConfig

	NodeUrl     string `json:"-"`
	MessagerUrl string `json:"-"`
	MarketUrl   string `json:"-"`
	GatewayUrl  string `json:"-"`
	AuthUrl     string `json:"-"`
	AuthToken   string `json:"-"`

	DbPlugin     string `json:"dbPlugin"`
	DbDns        string `json:"dbDns"`
	SendAddress  string `json:"sendAddress"`
	MinerAddress int64  `json:"minerAddress"`
}

type RenderParams struct {
	Config

	NameSpace       string
	PrivateRegistry string
	Args            []string

	TestID string
}

func DefaultConfig() Config {
	return Config{}
}

var PluginInfo = types.PluginInfo{
	Name:        "venus-sector-manager",
	Version:     version.Version(),
	Category:    types.Deploy,
	Repo:        "https://github.com/ipfs-force-community/venus-cluster.git",
	ImageTarget: "venus-sector-manager",
	Description: "",
}

var _ env.IDeployer = (*VenusSectorManagerDeployer)(nil)

type VenusSectorManagerDeployer struct {
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
	return &VenusSectorManagerDeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func NewVenusSectorManagerDeployer(env *env.K8sEnvDeployer, nodeUrl, messagerUrl, marketUrl, gatewayUrl, authUrl,
	authToken, dbPlugin, sendAddress string, minerAddress int64) *VenusSectorManagerDeployer {
	dbDns := ""
	if dbPlugin == "sqlxdb" {
		dbDns = env.FormatMysqlConnection("venus-sector-manager-" + env.UniqueId(""))
	}
	return &VenusSectorManagerDeployer{
		env: env,
		cfg: &Config{
			NodeUrl:      nodeUrl,
			MessagerUrl:  messagerUrl,
			MarketUrl:    marketUrl,
			GatewayUrl:   gatewayUrl,
			AuthUrl:      authUrl,
			AuthToken:    authToken,
			DbPlugin:     dbPlugin,
			DbDns:        dbDns,
			SendAddress:  sendAddress,
			MinerAddress: minerAddress,
		},
	}
}

func (deployer *VenusSectorManagerDeployer) Name() string {
	return PluginInfo.Name
}

func (deployer *VenusSectorManagerDeployer) Pods(ctx context.Context) ([]corev1.Pod, error) {
	return deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-sector-manager-%s-pod", deployer.env.UniqueId("")))
}

func (deployer *VenusSectorManagerDeployer) Deployment(ctx context.Context) ([]*appv1.Deployment, error) {
	return nil, nil
}

func (deployer *VenusSectorManagerDeployer) StatefulSet(ctx context.Context) (*appv1.StatefulSet, error) {
	return deployer.env.GetStatefulSet(ctx, deployer.statefulSetName)
}

func (deployer *VenusSectorManagerDeployer) Svc(ctx context.Context) (*corev1.Service, error) {
	return deployer.env.GetSvc(ctx, deployer.svcName)
}

func (deployer *VenusSectorManagerDeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

func (deployer *VenusSectorManagerDeployer) Param(key string) (interface{}, error) {
	return nil, errors.New("no params")
}

var f embed.FS

func (deployer *VenusSectorManagerDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		NameSpace:       deployer.env.NameSpace(),
		PrivateRegistry: deployer.env.PrivateRegistry(),
		TestID:          deployer.env.TestID(),
		Config:          *deployer.cfg,
	}

	// create configMap
	configMapFs, err := f.Open("venus-sector-manager/venus-sector-manager-configmap.yaml")
	if err != nil {
		return err
	}
	configMap, err := deployer.env.RunConfigMap(ctx, configMapFs, renderParams)
	if err != nil {
		return err
	}
	deployer.configMapName = configMap.GetName()

	// create deployment
	deployCfg, err := f.Open("venus-sector-manager/venus-sector-manager-statefulset.yaml")
	if err != nil {
		return err
	}
	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.statefulSetName = statefulSet.GetName()

	// create service
	svcCfg, err := f.Open("venus-sector-manager/venus-sector-manager-headless.yaml")
	svc, err := deployer.env.RunService(ctx, svcCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.configMapName = svc.GetName()

	deployer.svcEndpoint, err = deployer.env.WaitForServiceReady(ctx, deployer)
	if err != nil {
		return err
	}

	return nil
}

func (deployer *VenusSectorManagerDeployer) GetConfig(ctx context.Context) (interface{}, error) {
	cfgData, err := deployer.env.GetConfigMap(ctx, deployer.configMapName, "sector-manager.cfg")
	if err != nil {
		return nil, err
	}

	cfg := &modules.Config{}
	err = toml.Unmarshal(cfgData, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (deployer *VenusSectorManagerDeployer) Update(ctx context.Context, updateCfg interface{}) error {
	if updateCfg != nil {
		cfgData, err := toml.Marshal(updateCfg)
		if err != nil {
			return err
		}
		err = deployer.env.SetConfigMap(ctx, deployer.configMapName, "sector-manager.cfg", cfgData)
		if err != nil {
			return err
		}

		pods, err := deployer.Pods(ctx)
		if err != nil {
			return nil
		}
		for _, pod := range pods {
			_, err = deployer.env.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.venus-sector-manager/sector-manager.cfg")
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
