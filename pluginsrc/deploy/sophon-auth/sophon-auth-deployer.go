package sophonauth

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hunjixin/brightbird/types"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	"github.com/ipfs-force-community/sophon-auth/auth"
	"github.com/ipfs-force-community/sophon-auth/jwtclient"
	"github.com/pelletier/go-toml"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	env.BaseConfig

	MysqlDSN string `json:"-"`

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
		MysqlDSN: "",
	}
}

var PluginInfo = types.PluginInfo{
	Name:        "sophon-auth",
	Version:     version.Version(),
	PluginType:  types.Deploy,
	Repo:        "https://github.com/ipfs-force-community/sophon-auth.git",
	ImageTarget: "sophon-auth",
	Description: "",
}

var _ env.IDeployer = (*SophonAuthHASopDeployer)(nil)

type SophonAuthHASopDeployer struct { //nolint
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	configMapName   string
	statefulSetName string
	svcName         string

	params map[string]string
}

func DeployerFromConfig(envV *env.K8sEnvDeployer, cfg Config, params Config) (env.IDeployer, error) {
	defaultCfg := DefaultConfig()
	defaultCfg.MysqlDSN = envV.FormatMysqlConnection("sophon-auth-" + env.UniqueId(envV.TestID(), cfg.InstanceName))
	cfg, err := utils.MergeStructAndInterface(defaultCfg, cfg, params)
	if err != nil {
		return nil, err
	}
	return &SophonAuthHASopDeployer{
		env:    envV,
		cfg:    &cfg,
		params: make(map[string]string),
	}, nil
}

func DeployerFromBytes(env *env.K8sEnvDeployer, data json.RawMessage) (env.IDeployer, error) {
	cfg := &Config{}
	err := json.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}
	return &SophonAuthHASopDeployer{
		env:    env,
		cfg:    cfg,
		params: make(map[string]string),
	}, nil
}

func (deployer *SophonAuthHASopDeployer) InstanceName() (string, error) {
	return deployer.cfg.InstanceName, nil
}

func (deployer *SophonAuthHASopDeployer) Pods(ctx context.Context) ([]corev1.Pod, error) {
	return deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("sophon-auth-%s-pod", env.UniqueId(deployer.env.TestID(), deployer.cfg.InstanceName)))
}

func (deployer *SophonAuthHASopDeployer) StatefulSet(ctx context.Context) (*appv1.StatefulSet, error) {
	return deployer.env.GetStatefulSet(ctx, deployer.statefulSetName)
}

func (deployer *SophonAuthHASopDeployer) Svc(ctx context.Context) (*corev1.Service, error) {
	return deployer.env.GetSvc(ctx, deployer.svcName)
}

func (deployer *SophonAuthHASopDeployer) SvcEndpoint() (types.Endpoint, error) {
	return deployer.svcEndpoint, nil
}

func (deployer *SophonAuthHASopDeployer) Param(key string) (env.Params, error) {
	return env.ParamsFromVal(deployer.params[key]), nil
}

//go:embed sophon-auth
var f embed.FS

func (deployer *SophonAuthHASopDeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		NameSpace:       deployer.env.NameSpace(),
		PrivateRegistry: deployer.env.PrivateRegistry(),
		Args:            nil,
		UniqueId:        env.UniqueId(deployer.env.TestID(), deployer.cfg.InstanceName),
		Config:          *deployer.cfg,
	}

	//create database
	err = deployer.env.ResourceMgr().EnsureDatabase(deployer.cfg.MysqlDSN)
	if err != nil {
		return err
	}
	//create configmap
	configMapCfg, err := f.Open("sophon-auth/sophon-auth-configmap.yaml")
	if err != nil {
		return err
	}
	configMap, err := deployer.env.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.configMapName = configMap.GetName()

	//create deployment
	deployCfg, err := f.Open("sophon-auth/sophon-auth-statefulset.yaml")
	if err != nil {
		return err
	}
	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.statefulSetName = statefulSet.GetName()

	//create service
	svcCfg, err := f.Open("sophon-auth/sophon-auth-headless.yaml")
	if err != nil {
		return err
	}
	svc, err := deployer.env.RunService(ctx, svcCfg, renderParams)
	if err != nil {
		fmt.Println("service fail", err)
		return err
	}
	deployer.svcName = svc.GetName()

	deployer.svcEndpoint, err = deployer.env.WaitForServiceReady(ctx, deployer)
	if err != nil {

		fmt.Println("wait ready fail")
		return err
	}

	return deployer.prepareParams(ctx)
}

func (deployer *SophonAuthHASopDeployer) prepareParams(ctx context.Context) error {
	endpoint, err := deployer.SvcEndpoint()
	if err != nil {
		return err
	}
	sophonAuthPods, err := deployer.Pods(ctx)
	if err != nil {
		return err
	}

	svc, err := deployer.Svc(ctx)
	if err != nil {
		return err
	}
	if env.Debug {
		endpoint, err = deployer.env.PortForwardPod(ctx, sophonAuthPods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return err
		}
	}

	localToken, err := deployer.env.ReadSmallFilelInPod(ctx, sophonAuthPods[0].GetName(), "/root/.sophon-auth/token")
	if err != nil {
		return err
	}

	authAPIClient, err := jwtclient.NewAuthClient(endpoint.ToHTTP(), string(localToken))
	if err != nil {
		return err
	}

	_, err = authAPIClient.CreateUser(ctx, &auth.CreateUserRequest{
		Name:    "admin",
		Comment: utils.StringPtr("comment admin"),
		State:   0,
	})
	if err != nil && !strings.Contains(err.Error(), "user already exists") {
		return err
	}
	adminToken, err := authAPIClient.GenerateToken(ctx, "admin", "admin", "")
	if err != nil {
		return err
	}

	deployer.params["AdminToken"] = adminToken
	return nil
}

func (deployer *SophonAuthHASopDeployer) GetConfig(ctx context.Context) (env.Params, error) {
	cfgData, err := deployer.env.GetConfigMap(ctx, deployer.configMapName, "config.toml")
	if err != nil {
		return env.Params{}, err
	}

	return env.ParamsFromVal(cfgData), nil
}

func (deployer *SophonAuthHASopDeployer) Update(ctx context.Context, updateCfg interface{}) error {
	if updateCfg != nil {
		cfgData, err := toml.Marshal(updateCfg)
		if err != nil {
			return err
		}
		err = deployer.env.SetConfigMap(ctx, deployer.configMapName, "config.toml", cfgData)
		if err != nil {
			return err
		}

		pods, err := deployer.Pods(ctx)
		if err != nil {
			return nil
		}
		for _, pod := range pods {
			_, err = deployer.env.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.sophon-auth/config.toml")
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
