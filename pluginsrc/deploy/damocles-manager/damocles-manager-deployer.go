package damoclesmanager

import (
	"context"
	"embed"
	"fmt"
	"os"
	"strconv"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/abi"
	vapi "github.com/filecoin-project/venus/venus-shared/api"
	"github.com/ipfs-force-community/brightbird/env"
	venusutils "github.com/ipfs-force-community/brightbird/env/venus_utils"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/utils"
	"github.com/ipfs-force-community/brightbird/version"
	"github.com/ipfs-force-community/damocles/damocles-manager/core"
	"github.com/pelletier/go-toml"
	corev1 "k8s.io/api/core/v1"
)

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
make docker-push-manager TAG={{.Commit}} BUILD_DOCKER_PROXY={{.Proxy}} PRIVATE_REGISTRY={{.Registry}}`,
	},
	Description: "",
}

//go:embed damocles-manager
var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, cfg Config) (*DamoclesManagerReturn, error) {
	renderParams := RenderParams{
		NameSpace: k8sEnv.NameSpace(),
		Registry:  k8sEnv.Registry(),
		UniqueId:  env.UniqueId(k8sEnv.TestID(), cfg.InstanceName),
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
	statefulSet, err := k8sEnv.RunStatefulSets(ctx, deployCfg, renderParams)
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

func GetConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, configMapName string) (interface{}, error) {
	cfgData, err := k8sEnv.GetConfigMap(ctx, configMapName, "sector-manager.cfg")
	if err != nil {
		return nil, err
	}

	var cfg interface{}
	err = toml.Unmarshal(cfgData, &cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
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
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("damocles-manager-%s-pod", env.UniqueId(k8sEnv.TestID(), instanceName)))
}

func AddMiner(ctx context.Context, k8sEnv *env.K8sEnvDeployer, damoclesInstance DamoclesManagerReturn, minerCfg MinerCfg) error {
	//write config
	pods, err := GetPods(ctx, k8sEnv, damoclesInstance.InstanceName)
	if err != nil {
		return err
	}

	minerCfgFs, err := f.Open("damocles-manager/miner-cfg.toml")
	if err != nil {
		return err
	}

	data, err := env.QuickRender(minerCfgFs, struct {
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

	script := fmt.Sprintf(`cat <<EOT >> %s
	%s
EOT`, "/root/.damocles-manager/sector-manager.cfg", string(data))
	for _, pod := range pods {
		err = k8sEnv.ExecRemoteCmdWithStream(ctx, pod.Name, false, os.Stdout, nil, "/bin/bash", "-c", script)
		if err != nil {
			return err
		}
	}

	//restart
	for _, pod := range pods {
		err = k8sEnv.DeletePodAndWait(ctx, pod.GetName())
		if err != nil {
			return fmt.Errorf("restart pod %s", pod.GetName())
		}
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
