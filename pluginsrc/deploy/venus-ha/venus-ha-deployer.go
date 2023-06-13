package venus

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hunjixin/brightbird/types"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	env.BaseConfig

	AuthUrl        string   `json:"-"`
	AdminToken     string   `json:"-"`
	BootstrapPeers []string `json:"-"`

	NetType  string `json:"netType" description:"network type: mainnet,2k,calibrationnet,force"`
	Replicas int    `json:"replicas" description:"number of replicas"`
}

type RenderParams struct {
	Config

	NameSpace       string
	PrivateRegistry string
	Args            []string

	UniqueId string
}

func DefaultConfig() Config {
	return Config{
		Replicas: 1,
		NetType:  "force",
	}
}

var PluginInfo = types.PluginInfo{
	Name:        "venus-daemon-ha",
	Version:     version.Version(),
	PluginType:  types.Deploy,
	Repo:        "https://github.com/filecoin-project/venus.git",
	ImageTarget: "venus",
	Description: "",
}

var _ env.IDeployer = (*VenusHADeployer)(nil)

type VenusHADeployer struct { //nolint
	env *env.K8sEnvDeployer
	cfg *Config

	svcEndpoint types.Endpoint

	configMapName   string
	statefulSetName string
	svcName         string
}

func NewVenusHADeployer(env *env.K8sEnvDeployer, replicas int, authUrl string, adminToken string, bootstrapPeers ...string) *VenusHADeployer {
	return &VenusHADeployer{
		env: env,
		cfg: &Config{
			Replicas:       replicas, //default
			AuthUrl:        authUrl,
			AdminToken:     adminToken,
			BootstrapPeers: bootstrapPeers,
		},
	}
}

func DeployerFromConfig(env *env.K8sEnvDeployer, cfg Config, params Config) (env.IDeployer, error) {
	cfg, err := utils.MergeStructAndInterface(DefaultConfig(), cfg, params)
	if err != nil {
		return nil, err
	}
	return &VenusHADeployer{
		env: env,
		cfg: &cfg,
	}, nil
}

func (deployer *VenusHADeployer) InstanceName() (string, error) {
	return deployer.cfg.InstanceName, nil
}

func (deployer *VenusHADeployer) Pods(ctx context.Context) ([]corev1.Pod, error) {
	return deployer.env.GetPodsByLabel(ctx, fmt.Sprintf("venus-%s-pod", env.UniqueId(deployer.env.TestID(), deployer.cfg.InstanceName)))
}

func (deployer *VenusHADeployer) StatefulSet(ctx context.Context) (*appv1.StatefulSet, error) {
	return deployer.env.GetStatefulSet(ctx, deployer.statefulSetName)
}

func (deployer *VenusHADeployer) Svc(ctx context.Context) (*corev1.Service, error) {
	return deployer.env.GetSvc(ctx, deployer.svcName)
}

func (deployer *VenusHADeployer) SvcEndpoint() (types.Endpoint, error) {
	return deployer.svcEndpoint, nil
}

func (deployer *VenusHADeployer) Param(key string) (env.Params, error) {
	return env.Params{}, errors.New("no params")
}

//go:embed venus-node
var f embed.FS

func (deployer *VenusHADeployer) Deploy(ctx context.Context) (err error) {
	renderParams := RenderParams{
		NameSpace:       deployer.env.NameSpace(),
		PrivateRegistry: deployer.env.PrivateRegistry(),
		UniqueId:        env.UniqueId(deployer.env.TestID(), deployer.cfg.InstanceName),
		Args:            deployer.buildArgs(deployer.cfg.BootstrapPeers),
		Config:          *deployer.cfg,
	}

	//create configmap
	configMapCfg, err := f.Open("venus-node/venus-configmap.yaml")
	if err != nil {
		return err
	}
	fmt.Println(renderParams.BootstrapPeers)
	configMap, err := deployer.env.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.configMapName = configMap.GetName()

	//create statefulset
	deployCfg, err := f.Open("venus-node/venus-node-stateful-deployment.yaml")
	if err != nil {
		return err
	}
	statefulSet, err := deployer.env.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return err
	}
	deployer.statefulSetName = statefulSet.GetName()

	//create headless service
	svcCfg, err := f.Open("venus-node/venus-node-headless.yaml")
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

func (deployer *VenusHADeployer) GetConfig(ctx context.Context) (env.Params, error) {
	cfgData, err := deployer.env.GetConfigMap(ctx, deployer.configMapName, "config.json")
	if err != nil {
		return env.Params{}, err
	}

	return env.ParamsFromVal(cfgData), nil
}

// Update
// todo change this mode to config
func (deployer *VenusHADeployer) Update(ctx context.Context, updateCfg interface{}) error {
	if updateCfg != nil {
		cfgData, err := json.Marshal(updateCfg)
		if err != nil {
			return err
		}
		err = deployer.env.SetConfigMap(ctx, deployer.configMapName, "config.json", cfgData)
		if err != nil {
			return err
		}

		pods, err := deployer.Pods(ctx)
		if err != nil {
			return nil
		}
		for _, pod := range pods {
			_, err = deployer.env.ExecRemoteCmd(ctx, pod.GetName(), "echo", "'"+string(cfgData)+"'", ">", "/root/.venus/config.json")
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

func (deployer *VenusHADeployer) buildArgs(bootstrapPeers []string) []string {
	args := []string{
		"daemon",
		"--genesisfile=/shared-dir/devgen.car",
		"--import-snapshot=/shared-dir/dev-snapshot.car",
		"--network=" + deployer.cfg.NetType,
	}
	return args
}
