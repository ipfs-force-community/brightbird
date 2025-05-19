package damoclesmanager

import (
	"context"
	"embed"
	"fmt"
	"strconv"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/abi"
	vapi "github.com/filecoin-project/venus/venus-shared/api"
	"github.com/ipfs-force-community/damocles/damocles-manager/core"
	logging "github.com/ipfs/go-log/v2"
	"github.com/naoina/toml"
	corev1 "k8s.io/api/core/v1"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	venusutils "github.com/ipfs-force-community/brightbird/pkg/env/venus_utils"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/utils"
	"github.com/ipfs-force-community/brightbird/pkg/version"
)

var log = logging.Logger("damocles-manager")

type Config struct {
	env.BaseConfig
	VConfig
}

type VConfig struct {
	PieceStores   []string `jsonschema:"-"`
	PersistStores []string `jsonschema:"-"`

	NodeUrl     string `jsonschema:"-"`
	MessagerUrl string `jsonschema:"-"`
	MarketUrl   string `jsonschema:"-"`
	GatewayUrl  string `jsonschema:"-"`

	UserToken string `json:"userToken" jsonschema:"userToken" title:"UserToken" require:"true" `
}

type MinerCfg struct {
	SendFund            string          `json:"sendFund"  jsonschema:"sendFund" title:"sendFund" require:"true" default:"false" description:"sendFund"`
	MinerActor          abi.ActorID     `json:"minerActor"  jsonschema:"minerActor" title:"MinerActor" require:"true" `
	SenderWalletAddress address.Address `json:"senderWalletAddress"  jsonschema:"senderWalletAddress" title:"SenderWalletAddress" require:"true" `
}

type DamoclesManagerReturn struct { //nolint
	VConfig
	env.CommonDeployParams
}
type RenderParams struct {
	Config

	NameSpace string
	Registry  string
	Args      []string

	UniqueId      string
	MountStorages []string // distinct with PieceStores and PersistStores
}

var PluginInfo = types.PluginInfo{
	Name:       "damocles-manager",
	Version:    version.Version(),
	PluginType: types.Deploy,
	DeployPluginParams: types.DeployPluginParams{
		Repo:        "https://github.com/ipfs-force-community/damocles.git",
		ImageTarget: "damocles-manager",
		BuildScript: `make docker-push-manager TAG={{.Commit}} BUILD_DOCKER_PROXY={{.Proxy}} PRIVATE_REGISTRY={{.Registry}}`,
	},
	Description: "",
}

//go:embed damocles-manager
var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, cfg Config) (*DamoclesManagerReturn, error) {
	renderParams := RenderParams{
		NameSpace: k8sEnv.NameSpace(),
		Registry:  k8sEnv.Registry(),
		UniqueId:  env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), cfg.InstanceName),
		Config:    cfg,
	}

	if utils.HasDupItemInArrary(renderParams.PieceStores) {
		return nil, fmt.Errorf("piece storage has same pvc %s", renderParams.PieceStores)
	}

	if utils.HasDupItemInArrary(renderParams.PersistStores) {
		return nil, fmt.Errorf("piece storage has same pvc %s", renderParams.PersistStores)
	}

	pvcFilter := make(map[string]bool)
	mountStorages := []string{}
	for _, name := range renderParams.PieceStores {
		_, ok := pvcFilter[name]
		if !ok {
			pvcFilter[name] = true
			mountStorages = append(mountStorages, name)
		}
	}
	for _, name := range renderParams.PersistStores {
		_, ok := pvcFilter[name]
		if !ok {
			pvcFilter[name] = true
			mountStorages = append(mountStorages, name)
		}
	}
	renderParams.MountStorages = mountStorages

	// create configMap
	configMapFs, err := f.Open("damocles-manager/damocles-manager-configmap.yaml")
	if err != nil {
		return nil, err
	}
	configMap, err := k8sEnv.RunConfigMap(ctx, configMapFs, renderParams)
	if err != nil {
		return nil, err
	}

	// create deployment
	deployCfg, err := f.Open("damocles-manager/damocles-manager-statefulset.yaml")
	if err != nil {
		return nil, err
	}
	statefulSet, err := k8sEnv.RunStatefulSets(ctx, func(ctx context.Context, k8sEnv *env.K8sEnvDeployer) ([]corev1.Pod, error) {
		return GetPods(ctx, k8sEnv, cfg.InstanceName)
	}, deployCfg, renderParams)
	if err != nil {
		return nil, err
	}

	// create service
	svcCfg, err := f.Open("damocles-manager/damocles-manager-headless.yaml")
	if err != nil {
		return nil, err
	}

	svc, err := k8sEnv.RunService(ctx, svcCfg, renderParams)
	if err != nil {
		return nil, err
	}

	svcEndpoint, err := k8sEnv.WaitForServiceReady(ctx, svc, venusutils.VenusHealthCheck)
	if err != nil {
		return nil, err
	}

	return &DamoclesManagerReturn{
		VConfig: cfg.VConfig,
		CommonDeployParams: env.CommonDeployParams{
			BaseConfig:      cfg.BaseConfig,
			DeployName:      PluginInfo.Name,
			StatefulSetName: statefulSet.GetName(),
			ConfigMapName:   configMap.GetName(),
			SVCName:         svc.GetName(),
			SvcEndpoint:     svcEndpoint,
		},
	}, nil
}

func Update(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params DamoclesManagerReturn, updateCfg interface{}) error {
	cfgData, err := toml.Marshal(updateCfg)
	if err != nil {
		return err
	}
	err = k8sEnv.SetConfigMap(ctx, params.ConfigMapName, "sector-manager.cfg", cfgData)
	if err != nil {
		return err
	}

	pods, err := GetPods(ctx, k8sEnv, params.InstanceName)
	if err != nil {
		return nil
	}
	for _, pod := range pods {
		_, err = k8sEnv.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.damocles-manager/sector-manager.cfg")
		if err != nil {
			return err
		}
	}

	err = k8sEnv.UpdateStatefulSetsByName(ctx, params.StatefulSetName)
	if err != nil {
		return err
	}
	return nil
}

func GetPods(ctx context.Context, k8sEnv *env.K8sEnvDeployer, instanceName string) ([]corev1.Pod, error) {
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("damocles-manager-%s-pod", env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), instanceName)))
}

func AddMiner(ctx context.Context, k8sEnv *env.K8sEnvDeployer, damoclesInstance DamoclesManagerReturn, minerCfg MinerCfg) error {
	// 1. 获取configmap
	pods, err := GetPods(ctx, k8sEnv, damoclesInstance.InstanceName)
	if err != nil {
		return err
	}

	managerCfg, err := k8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), "cat", "/root/.damocles-manager/sector-manager.cfg")
	if err != nil {
		return err
	}
	log.Infoln("managerCfg is: ", string(managerCfg))

	minerCfgFs, err := f.Open("damocles-manager/miner-cfg.toml")
	if err != nil {
		return err
	}

	minerData, err := env.QuickRender(minerCfgFs, struct {
		SendFund            string
		MinerActor          string
		SenderWalletAddress string
	}{
		SendFund:            minerCfg.SendFund,
		MinerActor:          strconv.Itoa(int(minerCfg.MinerActor)),
		SenderWalletAddress: minerCfg.SenderWalletAddress.String(),
	})
	if err != nil {
		return err
	}

	// 2. 在sector-manager.cfg后面追加miner-cfg.toml
	managerCfg = append(managerCfg, minerData...)
	log.Infoln("managerCfg is: ", string(managerCfg))

	configmap, err := k8sEnv.GetConfigMapByName(ctx, damoclesInstance.ConfigMapName)
	if err != nil {
		return err
	}

	// 3. 使用新的sector-manager.cfg更新configmap
	configmap.Data["sector-manager.cfg"] = string(managerCfg)

	err = k8sEnv.UpdateConfigMaps(ctx, configmap)
	if err != nil {
		return err
	}

	statefulset, err := k8sEnv.GetStatefulSet(ctx, damoclesInstance.StatefulSetName)
	if err != nil {
		return err
	}

	err = k8sEnv.UpdateStatefulSets(ctx, statefulset)
	if err != nil {
		return err
	}

	svc, err := k8sEnv.GetSvc(ctx, damoclesInstance.SVCName)
	if err != nil {
		return err
	}

	_, err = k8sEnv.WaitForServiceReady(ctx, svc, venusutils.VenusHealthCheck)
	if err != nil {
		return err
	}

	return nil
}

func BuildDamoclesClient(ctx context.Context, damoclesManager DamoclesManagerReturn) (core.APIClient, jsonrpc.ClientCloser, error) {
	ainfo := vapi.NewAPIInfo(damoclesManager.SvcEndpoint.ToMultiAddr(), "")
	apiAddr, err := ainfo.DialArgs(vapi.VerString(core.MajorVersion))
	if err != nil {
		return core.APIClient{}, nil, err
	}

	var client core.APIClient
	closer, err := jsonrpc.NewMergeClient(ctx, apiAddr, core.APINamespace, []interface{}{&client}, ainfo.AuthHeader(), jsonrpc.WithRetry(true))
	if err != nil {
		return core.APIClient{}, nil, err
	}

	return client, closer, nil
}
