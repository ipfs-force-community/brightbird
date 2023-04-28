package venus_auth_ha

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/venus-auth/config"
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

	MysqlDSN string `json:"-"`

	Replicas int `json:"replicas"`
}

type RenderParams struct {
	Config

	PrivateRegistry string
	Args            []string
	UniqueId        string
}

func DefaultConfig() Config {
	return Config{
		Replicas: 1,
		MysqlDSN: "",
	}
}

var PluginInfo = types.PluginInfo{
	Name:        "venus-auth-ha",
	Version:     version.Version(),
	Category:    types.Deploy,
	Repo:        "https://github.com/filecoin-project/venus-auth.git",
	ImageTarget: "venus-auth",
	Description: "",
}

var _ env.IVenusAuthDeployer = (*VenusAuthHADeployer)(nil)

type VenusAuthHADeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	configMapName   string
	statefulSetName string
	svcName         string
}

func NewVenusAuthHADeployer(env *env.K8sEnvDeployer, replicas int) *VenusAuthHADeployer {
	return &VenusAuthHADeployer{
		env: env,
		cfg: &Config{
			Replicas: replicas, //default
			MysqlDSN: env.FormatMysqlConnection("venus-auth-ha-" + env.UniqueId("")),
		},
	}
}

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params Config) (env.IVenusAuthDeployer, error) {
	defaultCfg := DefaultConfig()
	defaultCfg.MysqlDSN = env.FormatMysqlConnection("venus-auth-ha-" + env.UniqueId(""))
	cfg, err := utils.MergeStructAndInterface(defaultCfg, cfg, params)
	if err != nil {
		return nil, err
	}
	return &VenusAuthHADeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func DeployerFromBytes(env *env.K8sEnvDeployer, data json.RawMessage) (env.IDeployer, error) {
	cfg := &Config{}
	err := json.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}
	return &VenusAuthHADeployer{
		env: env,
		cfg: cfg,
	}, nil
}

func (deployer *VenusAuthHADeployer) Name() string {
	return PluginInfo.Name
}

func (deployer *VenusAuthHADeployer) Pods(ctx context.Context) ([]corev1.Pod, error) {
	return deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-auth-%s-pod", deployer.env.UniqueId("")))
}

func (deployer *VenusAuthHADeployer) StatefulSet(ctx context.Context) (*appv1.StatefulSet, error) {
	return deployer.env.GetStatefulSet(ctx, deployer.statefulSetName)
}

func (deployer *VenusAuthHADeployer) Svc(ctx context.Context) (*corev1.Service, error) {
	return deployer.env.GetSvc(ctx, deployer.svcName)
}

func (deployer *VenusAuthHADeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

//go:embed venus-auth
var f embed.FS

func (deployer *VenusAuthHADeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		PrivateRegistry: deployer.env.PrivateRegistry(),
		Args:            nil,
		UniqueId:        deployer.env.UniqueId(""),
		Config:          *deployer.cfg,
	}

	//create database
	err = deployer.env.ResourceMgr().EnsureDatabase(deployer.cfg.MysqlDSN)
	if err != nil {
		return err
	}
	//create configmap
	configMapCfg, err := f.Open("venus-auth/venus-auth-configmap.yaml")
	if err != nil {
		return err
	}
	configMap, err := deployer.env.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.configMapName = configMap.GetName()

	//create deployment
	deployCfg, err := f.Open("venus-auth/venus-auth-ha-statefulset.yaml")
	if err != nil {
		return err
	}
	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.statefulSetName = statefulSet.GetName()

	//create service
	svcCfg, err := f.Open("venus-auth/venus-auth-headless.yaml")
	if err != nil {
		return err
	}
	svc, err := deployer.env.RunService(ctx, svcCfg, renderParams)
	if err != nil {
		fmt.Println("service fail", err)
		return err
	}
	deployer.svcName = svc.GetName()

	deployer.svcEndpoint, err = deployer.env.WaitForServiceReady(ctx, deployer)
	if err != nil {

		fmt.Println("wait ready fail")
		return err
	}
	return nil
}

func (deployer *VenusAuthHADeployer) GetConfig(ctx context.Context) (interface{}, error) {
	cfgData, err := deployer.env.GetConfigMap(ctx, deployer.configMapName, "config.toml")
	if err != nil {
		return nil, err
	}

	cfg := &config.Config{}
	err = toml.Unmarshal(cfgData, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (deployer *VenusAuthHADeployer) Update(ctx context.Context, updateCfg interface{}) error {
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
			_, err = deployer.env.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.venus-auth/config.toml")
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
