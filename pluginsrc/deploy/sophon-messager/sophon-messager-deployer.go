package venusmessager

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	"github.com/ipfs-force-community/sophon-messager/config"
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

	Replicas int `json:"replicas" description:"number of replicas"`
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
	Name:        "sophon-messager",
	Version:     version.Version(),
	PluginType:  types.Deploy,
	Repo:        "https://github.com/ipfs-force-community/sophon-messager.git",
	ImageTarget: "sophon-messager",
	Description: "",
}

var _ env.IDeployer = (*SophonMessagerDeployer)(nil)

type SophonMessagerDeployer struct { //nolint
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	pushPodName     string
	configMapName   string
	statefulSetName string
	svcName         string
}

func DeployerFromConfig(envV *env.K8sEnvDeployer, cfg Config, params Config) (env.IDeployer, error) {
	defaultCfg := DefaultConfig()
	defaultCfg.MysqlDSN = envV.FormatMysqlConnection("sophon-messager-ha-" + env.UniqueId(envV.TestID(), cfg.InstanceName))
	cfg, err := utils.MergeStructAndInterface(defaultCfg, cfg, params)
	if err != nil {
		return nil, err
	}
	return &SophonMessagerDeployer{
		env: envV,
		cfg: &cfg,
	}, nil
}

func (deployer *SophonMessagerDeployer) InstanceName() (string, error) {
	return deployer.cfg.InstanceName, nil
}

func (deployer *SophonMessagerDeployer) Pods(ctx context.Context) ([]corev1.Pod, error) {
	return deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("sophon-messager-%s-pod", env.UniqueId(deployer.env.TestID(), deployer.cfg.InstanceName)))
}

func (deployer *SophonMessagerDeployer) PushPods(ctx context.Context) ([]corev1.Pod, error) {
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

func (deployer *SophonMessagerDeployer) ReceivePods(ctx context.Context) ([]corev1.Pod, error) {
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

func (deployer *SophonMessagerDeployer) StatefulSet(ctx context.Context) (*appv1.StatefulSet, error) {
	return deployer.env.GetStatefulSet(ctx, deployer.statefulSetName)
}

func (deployer *SophonMessagerDeployer) Svc(ctx context.Context) (*corev1.Service, error) {
	return deployer.env.GetSvc(ctx, deployer.svcName)
}

func (deployer *SophonMessagerDeployer) SvcEndpoint() (types.Endpoint, error) {
	return deployer.svcEndpoint, nil
}

func (deployer *SophonMessagerDeployer) Param(key string) (env.Params, error) {
	return env.Params{}, errors.New("no params")
}

//go:embed sophon-messager
var f embed.FS

func (deployer *SophonMessagerDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		NameSpace:       deployer.env.NameSpace(),
		PrivateRegistry: deployer.env.PrivateRegistry(),
		UniqueId:        env.UniqueId(deployer.env.TestID(), deployer.cfg.InstanceName),
		Config:          *deployer.cfg,
	}

	//create database
	err = deployer.env.ResourceMgr().EnsureDatabase(renderParams.MysqlDSN)
	if err != nil {
		return err
	}

	//create configmap
	configMapCfg, err := f.Open("sophon-messager/sophon-messager-configmap.yaml")
	if err != nil {
		return err
	}
	configMap, err := deployer.env.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.configMapName = configMap.GetName()

	//deploy other node just service for others
	statefulSetCfg, err := f.Open("sophon-messager/sophon-messager-statefulset.yaml")
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
	_, err = deployer.env.ExecRemoteCmd(ctx, deployer.pushPodName, "sed", "-i", "-e", "'s/skipProcessHead = true/skipProcessHead = false/g'", "-e", "'s/skipPushMessage = true/skipPushMessage = false/g'", "/root/.sophon-messager/config.toml.tmp")
	if err != nil {
		return nil
	}

	log.Infof("change pod %s to a push node", deployer.pushPodName)
	//create service
	svcCfg, err := f.Open("sophon-messager/sophon-messager-headless.yaml")
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

func (deployer *SophonMessagerDeployer) GetConfig(ctx context.Context) (env.Params, error) {
	cfgData, err := deployer.env.ExecRemoteCmd(ctx, deployer.pushPodName, "cat /root/.sophon-messager/config.toml")
	if err != nil {
		return env.Params{}, err
	}

	return env.ParamsFromVal(cfgData), nil
}

func (deployer *SophonMessagerDeployer) Update(ctx context.Context, updateCfg interface{}) error {
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
			_, err = deployer.env.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.sophon-messager/config.toml")
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
