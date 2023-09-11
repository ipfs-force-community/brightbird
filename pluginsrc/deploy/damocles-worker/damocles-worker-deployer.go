package damoclesworker

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/utils"
	"github.com/ipfs-force-community/brightbird/version"
	logging "github.com/ipfs/go-log/v2"
	corev1 "k8s.io/api/core/v1"
)

var log = logging.Logger("messager-deployer")

type DamoclesWorkerConfig string //nolint

type Config struct {
	env.BaseConfig
	VConfig
}

type VConfig struct {
	SealPaths          []string `json:"sealPaths" jsonschema:"sealPaths" title:"Seal Path"  require:"true" `
	PieceStores        []string `jsonschema:"-"`
	PersistStores      []string `jsonschema:"-"`
	DamoclesManagerUrl string   `jsonschema:"-"`
	UserToken          string   `json:"userToken" jsonschema:"userToken" title:"User Token"  require:"true" `
	MinerAddress       string   `jsonschema:"-"`
}

type DropletMarketDeployReturn struct {
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
		BuildScript: `sed -i "2 i\RUN sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list" Dockerfile.manager
sed -i "17 i\RUN go env -w GOPROXY=https://goproxy.cn,direct" Dockerfile.manager
sed -i '7 i\ENV HTTPS_PROXY="{{.Proxy}}"' Dockerfile.manager
sed -i '8 i\ENV RUSTUP_DIST_SERVER="https://rsproxy.cn"' Dockerfile.manager
sed -i '9 i\ENV RUSTUP_UPDATE_ROOT="https://rsproxy.cn/rustup"' Dockerfile.manager
sed -i "s/https:\/\/sh.rustup.rs/https:\/\/rsproxy.cn\/rustup-init.sh/g" Dockerfile.manager

sed -i '1 i\export GITHUB_TOKEN={{.GitToken}}' Makefile
sed -i '2 i\export HTTPS_PROXY={{.Proxy}}' Makefile

cat > ./config << EOF
[source.crates-io]
replace-with = 'rsproxy'
[source.rsproxy]
registry = "https://rsproxy.cn/crates.io-index"
[source.rsproxy-sparse]
registry = "sparse+https://rsproxy.cn/index/"
[registries.rsproxy]
index = "https://rsproxy.cn/crates.io-index"
[net]
git-fetch-with-cli = true
EOF

sed -i "13 i\COPY ./config /usr/local/cargo/config" Dockerfile.manager

sed -i '5 i\export RUSTFLAGS=-C target-cpu=x86-64' damocles-worker/Makefile
sed -i '4 i\ENV HTTPS_PROXY="{{.Proxy}}"' damocles-worker/Dockerfile
sed -i "5 i\RUN sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list" damocles-worker/Dockerfile
cp config ./damocles-worker/config
sed -i "28 i\COPY ./config /usr/local/cargo/config" damocles-worker/Dockerfile
make docker-push TAG={{.Commit}} BUILD_DOCKER_PROXY={{.Proxy}} PRIVATE_REGISTRY={{.Registry}}`,
	},
	Description: "",
}

//go:embed  damocles-worker
var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, cfg Config) (*DropletMarketDeployReturn, error) {
	renderParams := RenderParams{
		NameSpace: k8sEnv.NameSpace(),
		Registry:  k8sEnv.Registry(),
		UniqueId:  env.UniqueId(k8sEnv.TestID(), cfg.InstanceName),
		Config:    cfg,
	}

	if utils.HasDupItemInArrary(renderParams.SealPaths) {
		return nil, fmt.Errorf("seal path not unique %s", renderParams.PieceStores)
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
	statefulSet, err := k8sEnv.RunStatefulSets(ctx, deployCfg, renderParams)
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

	return &DropletMarketDeployReturn{
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
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("damocles-worker-%s-pod", env.UniqueId(k8sEnv.TestID(), instanceName)))
}
