package sophonco

import (
	"context"
	"embed"
	"fmt"

	types2 "github.com/hunjixin/brightbird/types"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/version"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	env.BaseConfig
	VConfig
}
type VConfig struct {
	Replicas int `json:"replicas" description:"number of replicas"`

	AuthUrl    string   `ignore:"-" json:"authUrl"`
	AdminToken string   `ignore:"-" json:"adminToken"`
	Nodes      []string `ignore:"-" json:"nodes"`
}

type SophonCoDeployReturn struct {
	VConfig
	env.CommonDeployParams
}

type RenderParams struct {
	Config

	NameSpace       string
	PrivateRegistry string
	Args            []string

	UniqueId string
}

var PluginInfo = types2.PluginInfo{
	Name:        "sophon-co",
	Version:     version.Version(),
	PluginType:  types2.Deploy,
	Description: "",
	Repo:        "https://github.com/ipfs-force-community/sophon-co.git",
	ImageTarget: "sophon-co",
}

type SophonCoDeployer struct { //nolint
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types2.Endpoint

	statefulSetName string
	svcName         string
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

	statefulSet, err := k8sEnv.RunStatefulSets(ctx, deployCfg, renderParams)
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

	svcEndpoint, err := k8sEnv.WaitForServiceReady(ctx, svc)
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

	_, err = k8sEnv.RunStatefulSets(ctx, deployCfg, renderParams)
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
		NameSpace:       k8sEnv.NameSpace(),
		Config:          cfg,
		PrivateRegistry: k8sEnv.PrivateRegistry(),
		Args:            args,
		UniqueId:        env.UniqueId(k8sEnv.TestID(), cfg.InstanceName),
	}
}

func Pods(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params SophonCoDeployReturn) ([]corev1.Pod, error) {
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("sophon-co-%s-pod", env.UniqueId(k8sEnv.TestID(), params.InstanceName)))
}
