package damoclesworker

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	logging "github.com/ipfs/go-log/v2"
	corev1 "k8s.io/api/core/v1"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/utils"
	"github.com/ipfs-force-community/brightbird/pkg/version"
)

var log = logging.Logger("messager-deployer")

type DamoclesWorkerConfig string //nolint

type Config struct {
	env.BaseConfig
	VConfig
}

type VConfig struct {
	SealPaths          []string `json:"sealPaths" jsonschema:"sealPaths" title:"Seal Path"  require:"true" description:"seal path under path /seals/{.SsealPaths}" `
	PieceStores        []string `jsonschema:"-"`
	PersistStores      []string `jsonschema:"-"`
	DamoclesManagerUrl string   `jsonschema:"-"`
	UserToken          string   `json:"userToken" jsonschema:"userToken" title:"User Token"  require:"true" `
	MinerAddress       string   `jsonschema:"-"`
	DisableCC          bool     `json:"DisableCC" jsonschema:"DisableCC" title:"flag for disable_cc"  default:"false" require:"false" `
	EnableDeals        bool     `json:"EnableDeals" jsonschema:"EnableDeals" title:"flag for enable_deals"  default:"true" require:"false" `
}

type DamoclesWorkerReturn struct { //nolint
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
	Name:       "damocles-worker",
	Version:    version.Version(),
	PluginType: types.Deploy,
	DeployPluginParams: types.DeployPluginParams{
		Repo:        "https://github.com/ipfs-force-community/damocles.git",
		ImageTarget: "damocles-worker",
		BuildScript: `make docker-push-worker TAG={{.Commit}} BUILD_DOCKER_PROXY={{.Proxy}} PRIVATE_REGISTRY={{.Registry}}`,
	},
	Description: "",
}

//go:embed  damocles-worker
var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, cfg Config) (*DamoclesWorkerReturn, error) {
	renderParams := RenderParams{
		NameSpace: k8sEnv.NameSpace(),
		Registry:  k8sEnv.Registry(),
		UniqueId:  env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), cfg.InstanceName),
		Config:    cfg,
	}

	if utils.HasDupItemInArrary(renderParams.SealPaths) {
		return nil, fmt.Errorf("seal path not unique %s", renderParams.SealPaths)
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
	configMapCfg, err := f.Open("damocles-worker/damocles-worker-configmap.yaml")
	if err != nil {
		return nil, err
	}
	configMap, err := k8sEnv.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return nil, err
	}

	// create deployment
	deployCfg, err := f.Open("damocles-worker/damocles-worker-statefulset.yaml")
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
	svcCfg, err := f.Open("damocles-worker/damocles-worker-headless.yaml")
	if err != nil {
		return nil, err
	}

	svc, err := k8sEnv.RunService(ctx, svcCfg, renderParams)
	if err != nil {
		return nil, err
	}

	svcEndpoint, err := k8sEnv.WaitForServiceReady(ctx, svc, checkWorkerHealthy)
	if err != nil {
		return nil, err
	}

	return &DamoclesWorkerReturn{
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

func checkWorkerHealthy(ctx context.Context, endpoint types.Endpoint) error {
	buf := bytes.NewBufferString(`{"jsonrpc": "2.0", "method": "VenusWorker.WorkerVersion", "params": [], "id": 1}`)
	resp, err := retryablehttp.Post(endpoint.ToHTTP(), "application/json", buf)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	log.Debugf("track status %s %d", resp.Status, resp.StatusCode)
	return fmt.Errorf("receive health %s", resp.Status)
}

func GetPods(ctx context.Context, k8sEnv *env.K8sEnvDeployer, instanceName string) ([]corev1.Pod, error) {
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("damocles-worker-%s-pod", env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), instanceName)))
}
