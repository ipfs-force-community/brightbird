package venus_miner

import (
	"context"
	"embed"
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

	NodeUrl    string `json:"-"`
	AuthUrl    string `json:"-"`
	GatewayUrl string `json:"-"`
	AuthToken  string `json:"-"`

	UseMysql bool `json:"useMysql"`
}

type RenderParams struct {
	UniqueId string
	MysqlDSN string
	Config
}

func DefaultConfig() Config {
	return Config{}
}

var PluginInfo = types.PluginInfo{
	Name:        string(env.VenusMiner),
	Version:     version.Version(),
	Category:    types.Deploy,
	Description: "",
}

var _ env.IVenusMinerDeployer = (*VenusMinerDeployer)(nil)

type VenusMinerDeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	configMap  *corev1.ConfigMap
	pods       []corev1.Pod
	deployment []*appv1.Deployment
	svc        *corev1.Service
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

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params Config) (env.IDeployer, error) {
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

func (deployer *VenusMinerDeployer) Pods() []corev1.Pod {
	return deployer.pods
}

func (deployer *VenusMinerDeployer) Deployment() []*appv1.Deployment {
	return deployer.deployment
}

func (deployer *VenusMinerDeployer) Svc() *corev1.Service {
	return deployer.svc
}

func (deployer *VenusMinerDeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

//go:embed venus-miner
var f embed.FS

func (deployer *VenusMinerDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		UniqueId: deployer.env.UniqueId(""),
		Config:   *deployer.cfg,
	}
	if deployer.cfg.UseMysql {
		renderParams.MysqlDSN = "root:123456@tcp(192.168.200.103:3306)/venus-miner-" + deployer.env.UniqueId("") + "?charset=utf8&parseTime=True&loc=Local"
	}
	if len(renderParams.MysqlDSN) > 0 {
		deployer.env.CreateDatabase(renderParams.MysqlDSN)
	}
	//create configmap
	configMapCfg, err := f.Open("venus-miner/venus-miner-configmap.yaml")
	if err != nil {
		return err
	}
	deployer.configMap, err = deployer.env.CreateConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return err
	}

	//create deployment
	deployCfg, err := f.Open("venus-miner/venus-miner-deployment.yaml")
	if err != nil {
		return err
	}
	deployment, err := deployer.env.RunDeployment(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.deployment = append(deployer.deployment, deployment)

	deployer.pods, err = deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-miner-%s-pod", deployer.env.UniqueId("")))
	if err != nil {
		return err
	}

	//create service
	svcCfg, err := f.Open("venus-miner/venus-miner-service.yaml")
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
