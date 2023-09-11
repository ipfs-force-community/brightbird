package genesisnode

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/filecoin-project/go-address"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
	logging "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"
)

var log = logging.Logger("genesis-node")

type Config struct {
	env.BaseConfig
}

type RenderParams struct {
	NameSpace string
	Registry  string
	UniqueId  string
	Config
}

var PluginInfo = types.PluginInfo{
	Name:               "genesis-node",
	Version:            version.Version(),
	PluginType:         types.Deploy,
	DeployPluginParams: types.DeployPluginParams{}, //no build
	Description:        "create genesis node for test",
}

type GenesisReturn struct { //nolint
	Address        address.Address `json:"addr" jsonschema:"addr" title:"Blanance Account" description:"address used to get funds"`
	BootstrapPeer  string          `json:"bootstrapPeer" jsonschema:"bootstrapPeer" title:"Bootstrap Peer" description:"genesis node's ip endpoint"`
	RPCUrl         string          `json:"rpcUrl" jsonschema:"rpcUrl" title:"Rpc url" require:"true" description:"rpc url"`
	RPCToken       string          `json:"rpcToken" jsonschema:"rpcToken" title:"Token" require:"true" description:"rpc token"`
	GenesisStorage string          `json:"genesisStorage" jsonschema:"genesisStorage" title:"GenesisStorage" require:"true" description:"used to storeage devgen.car files"`
}

//go:embed genesis
var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, incomineCfg Config) (*GenesisReturn, error) {
	renderParams := RenderParams{
		NameSpace: k8sEnv.NameSpace(),
		Registry:  k8sEnv.Registry(),
		UniqueId:  env.UniqueId(k8sEnv.TestID(), incomineCfg.InstanceName),
		Config:    incomineCfg,
	}

	//create configmap
	configMapCfg, err := f.Open("genesis/genesis-configmap.yaml")
	if err != nil {
		return nil, err
	}

	_, err = k8sEnv.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create statefulset
	deployCfg, err := f.Open("genesis/genesis-stateful-deployment.yaml")
	if err != nil {
		return nil, err
	}
	_, err = k8sEnv.RunStatefulSets(ctx, deployCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create headless service
	svcCfg, err := f.Open("genesis/genesis-headless.yaml")
	if err != nil {
		return nil, err
	}
	svc, err := k8sEnv.RunService(ctx, svcCfg, renderParams)
	if err != nil {
		return nil, err
	}

	svcEndpoint, err := k8sEnv.WaitForServiceReady(ctx, svc, checkLotusHealthy)
	if err != nil {
		return nil, err
	}

	pods, err := k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("genesis-%s-pod", env.UniqueId(k8sEnv.TestID(), incomineCfg.InstanceName)))
	if err != nil {
		return nil, err
	}
	// /lotus wallet import --as-default ~/.genesis-sectors/pre-seal-t01000.key ;
	// /lotus-miner init --genesis-miner --actor=t01000 --sector-size=2KiB --pre-sealed-sectors=/root/genesis/.genesis-sectors --pre-sealed-metadata=/root/genesis/.genesis-sectors/pre-seal-t01000.json --nosync ;
	// /lotus-miner run --nosync ;
	importResult, err := k8sEnv.ExecRemoteCmd(ctx, pods[0].Name, "/lotus", "wallet", "import", "--as-default", "/root/genesis/.genesis-sectors/pre-seal-t01000.key")
	if err != nil {
		return nil, fmt.Errorf("import key fail %w", err)
	}
	// imported key t3tehwiess4l72p5rfz6rzppx42kcp25clcxhz6mvjghhy6ulqtrom24t5tkarr443lx3e2sso6j7i7d6g6poa successfully!
	seq := strings.Split(string(importResult), " ")
	addr, err := address.NewFromString(strings.Trim(seq[2], " \t\r"))
	if err != nil {
		return nil, err
	}

	err = k8sEnv.ExecRemoteCmdWithStream(ctx, pods[0].Name, true, os.Stdout, nil, "/lotus-miner", "init", "--genesis-miner", "--actor=t01000", "--sector-size=2KiB", "--pre-sealed-sectors=/root/genesis/.genesis-sectors", "--pre-sealed-metadata=/root/genesis/.genesis-sectors/pre-seal-t01000.json", "--nosync")
	if err != nil {
		return nil, fmt.Errorf("init genesis miner fail %w", err)
	}

	log.Infof("run lotus-miner background")

	err = k8sEnv.ExecRemoteCmdWithStream(ctx, pods[0].Name, false, os.Stdout, nil, "/bin/bash", "-c", "nohup /lotus-miner run --nosync > /root/lotus-miner.log 2>&1 &")
	if err != nil {
		return nil, fmt.Errorf("run lotus-miner fail %w", err)
	}

	err = k8sEnv.ExecRemoteCmdWithStream(ctx, pods[0].Name, true, os.Stdout, nil, "/bin/bash", "-c", "ps x|grep lotus-miner")
	if err != nil {
		return nil, fmt.Errorf("run lotus-miner fail %w", err)
	}

	token, err := k8sEnv.ReadSmallFilelInPod(ctx, pods[0].Name, "/root/.lotus/token")
	if err != nil {
		return nil, err
	}

	libP2pArr, err := k8sEnv.ExecRemoteCmd(ctx, pods[0].Name, "/lotus", "net", "listen")
	if err != nil {
		return nil, fmt.Errorf("exec net listen fail %w", err)
	}

	libP2p := strings.Trim(strings.Split(string(libP2pArr), "\n")[0], " \t\r")
	mr, err := ma.NewMultiaddr(libP2p)
	if err != nil {
		return nil, fmt.Errorf("parser libp2p %s  %w", libP2p, err)
	}

	port, err := mr.ValueForProtocol(ma.P_TCP)
	if err != nil {
		return nil, fmt.Errorf("unable to get p2p port %w", err)
	}

	peer, err := mr.ValueForProtocol(ma.P_P2P)
	if err != nil {
		return nil, fmt.Errorf("unable to get p2p peer %w", err)
	}

	pod, err := k8sEnv.GetPod(ctx, pods[0].GetName())
	if err != nil {
		return nil, fmt.Errorf("unable to get pod %w", err)
	}

	claimName := ""
	for _, vol := range pod.Spec.Volumes {
		if vol.Name == "genesis-pvc" {
			claimName = vol.PersistentVolumeClaim.ClaimName
		}
	}
	log.Infof("genesis pvc name  %s", claimName)

	return &GenesisReturn{
		Address:        addr,
		GenesisStorage: claimName,
		BootstrapPeer:  fmt.Sprintf("/dns/%s/tcp/%s/p2p/%s", svcEndpoint.IP(), port, peer),
		RPCUrl:         svcEndpoint.ToMultiAddr(),
		RPCToken:       string(token),
	}, nil
}

func checkLotusHealthy(_ context.Context, endpoint types.Endpoint) error {
	log.Infof("try to check health at %s", endpoint.ToHTTP())
	resp, err := retryablehttp.Get(fmt.Sprintf("%s/health/livez", endpoint.ToHTTP()))
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("receive health %s", resp.Status)
}
