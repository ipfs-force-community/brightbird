package sophonauth

import (
	"context"
	"embed"
	"fmt"
	"strings"

	"github.com/ipfs-force-community/brightbird/types"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/utils"
	"github.com/ipfs-force-community/brightbird/version"
	"github.com/ipfs-force-community/sophon-auth/auth"

	"github.com/ipfs-force-community/sophon-auth/config"
	"github.com/ipfs-force-community/sophon-auth/jwtclient"
	"github.com/pelletier/go-toml"
)

type Config struct {
	env.BaseConfig

	MysqlDSN string `jsonschema:"-" json:"mysqlDSN"`

	Replicas int `json:"replicas" jsonschema:"replicas" title:"replicas" default:"1" require:"true" description:"number of replicas"`
}

func DefaultConfig() Config {
	return Config{
		Replicas: 1,
		MysqlDSN: "",
	}
}

type RenderParams struct {
	Config

	NameSpace       string
	PrivateRegistry string
	Args            []string
	UniqueId        string
}

type SophonAuthDeployReturn struct { //nolint
	MysqlDSN   string `json:"mysqlDSN"`
	Replicas   int    `json:"replicas" description:"number of replicas"`
	AdminToken string `json:"adminToken"`
	env.CommonDeployParams
}

var PluginInfo = types.PluginInfo{
	Name:        "sophon-auth",
	Version:     version.Version(),
	PluginType:  types.Deploy,
	Repo:        "https://github.com/ipfs-force-community/sophon-auth.git",
	ImageTarget: "sophon-auth",
	Description: "",
}

//go:embed sophon-auth
var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, cfg Config) (*SophonAuthDeployReturn, error) {
	cfg.MysqlDSN = k8sEnv.FormatMysqlConnection("sophon-auth-" + env.UniqueId(k8sEnv.TestID(), cfg.InstanceName))
	renderParams := RenderParams{
		NameSpace:       k8sEnv.NameSpace(),
		PrivateRegistry: k8sEnv.PrivateRegistry(),
		Args:            nil,
		UniqueId:        env.UniqueId(k8sEnv.TestID(), cfg.InstanceName),
		Config:          cfg,
	}

	//create database
	err := k8sEnv.ResourceMgr().EnsureDatabase(cfg.MysqlDSN)
	if err != nil {
		return nil, err
	}
	//create configmap
	configMapCfg, err := f.Open("sophon-auth/sophon-auth-configmap.yaml")
	if err != nil {
		return nil, err
	}
	configMap, err := k8sEnv.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create deployment
	deployCfg, err := f.Open("sophon-auth/sophon-auth-statefulset.yaml")
	if err != nil {
		return nil, err
	}
	statefulSet, err := k8sEnv.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create service
	svcCfg, err := f.Open("sophon-auth/sophon-auth-headless.yaml")
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

	adminToken, err := GenerateAdminToken(ctx, k8sEnv, cfg.InstanceName, svcEndpoint)
	if err != nil {
		return nil, err
	}

	return &SophonAuthDeployReturn{
		MysqlDSN:   cfg.MysqlDSN,
		Replicas:   cfg.Replicas,
		AdminToken: adminToken,
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

func GenerateAdminToken(ctx context.Context, k8sEnv *env.K8sEnvDeployer, isntanceName string, endpoint types.Endpoint) (string, error) {
	pods, err := k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("sophon-auth-%s-pod", env.UniqueId(k8sEnv.TestID(), isntanceName)))
	if err != nil {
		return "", err
	}

	localToken, err := k8sEnv.ReadSmallFilelInPod(ctx, pods[0].GetName(), "/root/.sophon-auth/token")
	if err != nil {
		return "", err
	}

	authAPIClient, err := jwtclient.NewAuthClient(endpoint.ToHTTP(), string(localToken))
	if err != nil {
		return "", err
	}

	_, err = authAPIClient.CreateUser(ctx, &auth.CreateUserRequest{
		Name:    "admin",
		Comment: utils.StringPtr("comment admin"),
		State:   0,
	})
	if err != nil && !strings.Contains(err.Error(), "user already exists") {
		return "", err
	}
	adminToken, err := authAPIClient.GenerateToken(ctx, "admin", "admin", "")
	if err != nil {
		return "", err
	}

	return adminToken, nil
}

func GetConfig(ctx context.Context, envCtx *env.K8sEnvDeployer, configMapName string) (config.Config, error) {
	cfgData, err := envCtx.GetConfigMap(ctx, configMapName, "config.toml")
	if err != nil {
		return config.Config{}, err
	}

	var cfg config.Config
	err = toml.Unmarshal(cfgData, &cfg)
	if err != nil {
		return config.Config{}, err
	}

	return cfg, nil
}

func Update(ctx context.Context, k8sEnv *env.K8sEnvDeployer, deployParams SophonAuthDeployReturn, updateCfg config.Config) error {
	cfgData, err := toml.Marshal(updateCfg)
	if err != nil {
		return err
	}

	err = k8sEnv.SetConfigMap(ctx, deployParams.ConfigMapName, "config.toml", cfgData)
	if err != nil {
		return err
	}

	pods, err := k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("sophon-auth-%s-pod", env.UniqueId(k8sEnv.TestID(), deployParams.InstanceName)))
	if err != nil {
		return err
	}

	for _, pod := range pods {
		_, err = k8sEnv.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.sophon-auth/config.toml")
		if err != nil {
			return err
		}
	}

	err = k8sEnv.UpdateStatefulSets(ctx, deployParams.StatefulSetName)
	if err != nil {
		return err
	}
	return nil
}
