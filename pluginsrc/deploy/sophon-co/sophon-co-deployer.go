package sophonco

import (
	"context"
	"embed"
	"fmt"

	venusutils "github.com/ipfs-force-community/brightbird/env/venus_utils"
	types2 "github.com/ipfs-force-community/brightbird/types"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/version"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	env.BaseConfig
	VConfig
}
type VConfig struct {
	Replicas int `json:"replicas"  jsonschema:"replicas" title:"replicas" default:"1" require:"true" description:"number of replicas"`

	AuthUrl    string   `jsonschema:"-" json:"authUrl"`
	AdminToken string   `jsonschema:"-" json:"adminToken"`
	Nodes      []string `jsonschema:"-" json:"nodes"`
}

type SophonCoDeployReturn struct { //nolint
	VConfig
	env.CommonDeployParams
}

type RenderParams struct {
	Config

	NameSpace string
	Registry  string
	Args      []string

	UniqueId string
}

var PluginInfo = types2.PluginInfo{
	Name:        "sophon-co",
	Version:     version.Version(),
	PluginType:  types2.Deploy,
	Description: "",
	DeployPluginParams: types2.DeployPluginParams{
		Repo:        "https://github.com/ipfs-force-community/sophon-co.git",
		ImageTarget: "sophon-co",
		BuildScript: `make docker-push TAG={{.Commit}} BUILD_DOCKER_PROXY={{.Proxy}} PRIVATE_REGISTRY={{.Registry}}`,
	},
}

//go:embed  sophon-co
var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, cfg Config) (*SophonCoDeployReturn, error) {
	renderParams := buildRenderParams(k8sEnv, cfg)

	//create deployment
	deployCfg, err := f.Open("sophon-co/sophon-co-statefulset.yaml")
	if err != nil {
		return nil, err
	}

	statefulSet, err := k8sEnv.RunStatefulSets(ctx, func(ctx context.Context, k8sEnv *env.K8sEnvDeployer) ([]corev1.Pod, error) {
		return GetPods(ctx, k8sEnv, cfg.InstanceName)
	}, deployCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create service
	svcCfg, err := f.Open("sophon-co/sophon-co-headless.yaml")
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
	return &SophonCoDeployReturn{
		VConfig: cfg.VConfig,
		CommonDeployParams: env.CommonDeployParams{
			BaseConfig:      cfg.BaseConfig,
			DeployName:      PluginInfo.Name,
			StatefulSetName: statefulSet.GetName(),
			ConfigMapName:   "",
			SVCName:         svc.GetName(),
			SvcEndpoint:     svcEndpoint,
		},
	}, nil
}

func Update(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params *Config) error {
	//restart
	renderParams := buildRenderParams(k8sEnv, *params)
	// create deployment
	deployCfg, err := f.Open("sophon-co/sophon-co-statefulset.yaml")
	if err != nil {
		return err
	}

	_, err = k8sEnv.RunStatefulSets(ctx, func(ctx context.Context, k8sEnv *env.K8sEnvDeployer) ([]corev1.Pod, error) {
		return GetPods(ctx, k8sEnv, params.InstanceName)
	}, deployCfg, renderParams)
	return err
}

func buildRenderParams(k8sEnv *env.K8sEnvDeployer, cfg Config) RenderParams {
	var args []string
	for _, node := range cfg.Nodes {
		args = append(args, "--node")
		args = append(args, node)
	}

	args = append(args, "--auth")
	args = append(args, cfg.AdminToken+":"+cfg.AuthUrl)

	return RenderParams{
		NameSpace: k8sEnv.NameSpace(),
		Config:    cfg,
		Registry:  k8sEnv.Registry(),
		Args:      args,
		UniqueId:  env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), cfg.InstanceName),
	}
}

func GetPods(ctx context.Context, k8sEnv *env.K8sEnvDeployer, instanceName string) ([]corev1.Pod, error) {
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("sophon-co-%s-pod", env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), instanceName)))
}
