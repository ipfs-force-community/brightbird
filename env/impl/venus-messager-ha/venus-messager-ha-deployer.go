package venus_messager_ha

import (
	"context"
	"embed"
	"fmt"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	logging "github.com/ipfs/go-log/v2"
	"github.com/pelletier/go-toml"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var log = logging.Logger("messager-deployer")

type Config struct {
	env.BaseConfig

	NodeUrl    string `json:"-"`
	GatewayUrl string `json:"-"`
	AuthUrl    string `json:"-"`
	AuthToken  string `json:"-"`
	MysqlDSN   string `json:"-"`

	Replicas int `json:"replicas"`
}

type RenderParams struct {
	Config

	NameSpace       string
	PrivateRegistry string
	Args            []string
	UniqueId        string
}

func DefaultConfig() Config {
	return Config{
		Replicas: 1,
	}
}

var PluginInfo = types.PluginInfo{
	Name:        "venus-message-ha",
	Version:     version.Version(),
	Category:    types.Deploy,
	Repo:        "https://github.com/filecoin-project/venus-messager.git",
	ImageTarget: "venus-messager",
	Description: "",
}

var _ env.IDeployer = (*VenusMessagerHADeployer)(nil)

type VenusMessagerHADeployer struct {
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	pushPodName     string
	configMapName   string
	statefulSetName string
	svcName         string
}

func NewVenusMessagerHADeployer(env *env.K8sEnvDeployer, replicas int, nodeUrl, authUrl, gatewayUrl, authToken string) *VenusMessagerHADeployer {
	return &VenusMessagerHADeployer{
		env: env,
		cfg: &Config{
			Replicas:   replicas, //default
			AuthUrl:    authUrl,
			GatewayUrl: gatewayUrl,
			AuthToken:  authToken,
			NodeUrl:    nodeUrl,
			MysqlDSN:   env.FormatMysqlConnection("venus-messager-ha-" + env.UniqueId("")),
		},
	}
}

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params Config) (env.IDeployer, error) {
	defaultCfg := DefaultConfig()
	defaultCfg.MysqlDSN = env.FormatMysqlConnection("venus-messager-ha-" + env.UniqueId(""))
	cfg, err := utils.MergeStructAndInterface(defaultCfg, cfg, params)
	if err != nil {
		return nil, err
	}
	return &VenusMessagerHADeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func (deployer *VenusMessagerHADeployer) Name() string {
	return PluginInfo.Name
}

func (deployer *VenusMessagerHADeployer) Pods(ctx context.Context) ([]corev1.Pod, error) {
	return deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-messager-%s-pod", deployer.env.UniqueId("")))
}

func (deployer *VenusMessagerHADeployer) PushPods(ctx context.Context) ([]corev1.Pod, error) {
	pods, err := deployer.Pods(ctx)
	if err != nil {
		return nil, err
	}
	var pushPods []corev1.Pod
	for _, pod := range pods {
		if pod.Labels["rule"] == "push" {
			pushPods = append(pushPods, pod)
		}
	}
	return pushPods, nil
}

func (deployer *VenusMessagerHADeployer) ReceivePods(ctx context.Context) ([]corev1.Pod, error) {
	pods, err := deployer.Pods(ctx)
	if err != nil {
		return nil, err
	}
	var pushPods []corev1.Pod
	for _, pod := range pods {
		if pod.Labels["rule"] == "receive" {
			pushPods = append(pushPods, pod)
		}
	}
	return pushPods, nil
}

func (deployer *VenusMessagerHADeployer) StatefulSet(ctx context.Context) (*appv1.StatefulSet, error) {
	return deployer.env.GetStatefulSet(ctx, deployer.statefulSetName)
}

func (deployer *VenusMessagerHADeployer) Svc(ctx context.Context) (*corev1.Service, error) {
	return deployer.env.GetSvc(ctx, deployer.svcName)
}

func (deployer *VenusMessagerHADeployer) SvcEndpoint() types.Endpoint {
	return deployer.svcEndpoint
}

//go:embed venus-messager
var f embed.FS

func (deployer *VenusMessagerHADeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		NameSpace:       deployer.env.NameSpace(),
		PrivateRegistry: deployer.env.PrivateRegistry(),
		UniqueId:        deployer.env.UniqueId(""),
		Config:          *deployer.cfg,
	}

	//create database
	err = deployer.env.ResourceMgr().EnsureDatabase(renderParams.MysqlDSN)
	if err != nil {
		return err
	}

	//create configmap
	configMapCfg, err := f.Open("venus-messager/venus-messager-configmap.yaml")
	if err != nil {
		return err
	}
	configMap, err := deployer.env.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.configMapName = configMap.GetName()

	//deploy other node just service for others
	statefulSetCfg, err := f.Open("venus-messager/venus-messager-statefulset.yaml")
	if err != nil {
		return err
	}
	statefulSet, err := deployer.env.RunStatefulSets(ctx, statefulSetCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.statefulSetName = statefulSet.GetName()

	//change the first node to a push node
	pods, err := deployer.Pods(ctx)
	if err != nil {
		return nil
	}
	deployer.pushPodName = pods[0].GetName()
	_, err = deployer.env.ExecRemoteCmd(ctx, deployer.pushPodName, "sed", "-i", "-e", "'s/skipProcessHead = true/skipProcessHead = false/g'", "-e", "'s/skipPushMessage = true/skipPushMessage = false/g'", "/root/.venus-messager/config.toml.tmp")
	if err != nil {
		return nil
	}

	log.Infof("change pod %s to a push node", deployer.pushPodName)
	//create service
	svcCfg, err := f.Open("venus-messager/venus-messager-headless.yaml")
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

func (deployer *VenusMessagerHADeployer) GetConfig(ctx context.Context) (interface{}, error) {
	cfgData, err := deployer.env.ExecRemoteCmd(ctx, deployer.pushPodName, "cat /root/.venus-messager/config.toml")
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

func (deployer *VenusMessagerHADeployer) Update(ctx context.Context, updateCfg interface{}) error {
	if updateCfg != nil {
		update := updateCfg.(*config.Config)
		pods, err := deployer.Pods(ctx)
		if err != nil {
			return nil
		}
		for _, pod := range pods {
			if pod.GetName() == deployer.pushPodName {
				update.MessageService.SkipProcessHead = false
				update.MessageService.SkipPushMessage = false
			} else {

				update.MessageService.SkipProcessHead = true
				update.MessageService.SkipPushMessage = true
			}

			cfgData, err := toml.Marshal(update)
			if err != nil {
				return err
			}
			_, err = deployer.env.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.venus-messager/config.toml")
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
