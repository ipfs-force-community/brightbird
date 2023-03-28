package venus_auth_ha

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
	env.BaseConfig

	MysqlDSN string `json:"-"`

	Replicas int `json:"replicas"`
}

type RenderParams struct {
	UniqueId string
	Config
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

	pods       []corev1.Pod
	deployment []*appv1.Deployment
	svc        *corev1.Service
}

func NewVenusAuthHADeployer(env *env.K8sEnvDeployer, replicas int) *VenusAuthHADeployer {
	return &VenusAuthHADeployer{
		env: env,
		cfg: &Config{
			Replicas: replicas, //default
			MysqlDSN: "root:123456@tcp(192.168.200.103:3306)/venus-auth-" + env.UniqueId("") + "?charset=utf8&parseTime=True&loc=Local",
		},
	}
}

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params Config) (env.IDeployer, error) {
	defaultCfg := DefaultConfig()
	defaultCfg.MysqlDSN = "root:123456@tcp(192.168.200.103:3306)/venus-auth-" + env.UniqueId("") + "?charset=utf8&parseTime=True&loc=Local"
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

func (deployer *VenusAuthHADeployer) Pods() []corev1.Pod {
	return deployer.pods
}

func (deployer *VenusAuthHADeployer) Deployment() []*appv1.Deployment {
	return deployer.deployment
}

func (deployer *VenusAuthHADeployer) Svc() *corev1.Service {
	return deployer.svc
}

func (deployer *VenusAuthHADeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

//go:embed venus-auth
var f embed.FS

func (deployer *VenusAuthHADeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		UniqueId: deployer.env.UniqueId(""),
		Config:   *deployer.cfg,
	}

	//create database
	err = deployer.env.CreateDatabase(deployer.cfg.MysqlDSN)
	if err != nil {
		return err
	}

	//create deployment
	deployCfg, err := f.Open("venus-auth/venus-auth-ha-deployment.yaml")
	if err != nil {
		return err
	}
	deployment, err := deployer.env.RunDeployment(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.deployment = append(deployer.deployment, deployment)

	deployer.pods, err = deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-auth-%s-pod", deployer.env.UniqueId("")))
	if err != nil {
		return err
	}

	//create service
	svcCfg, err := f.Open("venus-auth/venus-auth-service.yaml")
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
