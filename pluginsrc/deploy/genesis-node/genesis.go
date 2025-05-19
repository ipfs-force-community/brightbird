package genesisnode

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/hashicorp/go-retryablehttp"
	logging "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"
	corev1 "k8s.io/api/core/v1"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"
)

var log = logging.Logger("genesis-node")

const (
	vrkAddr, vrkInfo      = "f3v5dbvxeibanltz7jfrcsm76ixkps4exdt3g4uwfti2vte5btcwb5joi47yz5lehm6wbzdrcznjoff33zmbxa", "7b2254797065223a22626c73222c22507269766174654b6579223a225635754e4a6d727244626f2f633054685a563276426538536b4a636b4470444c724951316d33725744676f3d227d"
	notaryAddr, notaryKey = "f3qrmprkwyononbzfxlwdxrauj3bek5rwwrkq46n7egdl6rabiua6pdjgnw5snxlppip67aitnnt3iw2wtvleq", "7b2254797065223a22626c73222c22507269766174654b6579223a2258787131573365357a704e4e4d6e424a75425538332f3376336e73684978446c306a4b466264556a5453303d227d"
)

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
	Name:       "genesis-node",
	Version:    version.Version(),
	PluginType: types.Deploy,
	DeployPluginParams: types.DeployPluginParams{
		Repo:        "https://github.com/ipfs-force-community/lotus.git",
		ImageTarget: "lotus",
		BuildScript: `make docker-push TAG={{.Commit}} BUILD_DOCKER_PROXY={{.Proxy}} PRIVATE_REGISTRY={{.Registry}}`,
	},
	Description: "create genesis node for test",
}

type GenesisReturn struct { //nolint
	env.CommonDeployParams

	Address        address.Address `json:"addr" jsonschema:"addr" title:"Blanance Account" description:"address used to get funds"`
	BootstrapPeers []string        `json:"bootstrapPeers" jsonschema:"bootstrapPeers" title:"Bootstrap Peer" description:"genesis node's ip endpoint"`
	RPCUrl         string          `json:"rpcUrl" jsonschema:"rpcUrl" title:"Rpc url" require:"true" description:"rpc url"`
	RPCToken       string          `json:"rpcToken" jsonschema:"rpcToken" title:"Token" require:"true" description:"rpc token"`
	GenesisStorage string          `json:"genesisStorage" jsonschema:"genesisStorage" title:"GenesisStorage" require:"true" description:"used to storeage devgen.car files"`
	F01000Key      string          `json:"f01000Key" jsonschema:"f01000Key" title:"f01000-private-key" description:"private key for f01000"`

	VrkAddr    address.Address `json:"vrkAddr" jsonschema:"vrkAddr" title:"vrk-address"  description:"verified registry's root address"`
	NotaryAddr address.Address `json:"notaryAddr" jsonschema:"notaryAddr" title:"notary-address"  description:"notary address"`
}

//go:embed genesis
var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, incomineCfg Config) (*GenesisReturn, error) {
	renderParams := RenderParams{
		NameSpace: k8sEnv.NameSpace(),
		Registry:  k8sEnv.Registry(),
		UniqueId:  env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), incomineCfg.InstanceName),
		Config:    incomineCfg,
	}

	//create configmap
	configMapCfg, err := f.Open("genesis/genesis-configmap.yaml")
	if err != nil {
		return nil, err
	}

	configMap, err := k8sEnv.RunConfigMap(ctx, configMapCfg, renderParams)
	if err != nil {
		return nil, err
	}

	//create statefulset
	deployCfg, err := f.Open("genesis/genesis-stateful-deployment.yaml")
	if err != nil {
		return nil, err
	}
	statefulSet, err := k8sEnv.RunStatefulSets(ctx, func(ctx context.Context, k8sEnv *env.K8sEnvDeployer) ([]corev1.Pod, error) {
		return GetPods(ctx, k8sEnv, incomineCfg.InstanceName)
	}, deployCfg, renderParams)
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
	pods, err := k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("genesis-%s-pod", env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), incomineCfg.InstanceName)))
	if err != nil {
		return nil, err
	}
	if len(pods) == 0 {
		return nil, fmt.Errorf("no pod found")
	}
	execPodName := pods[0].Name
	log.Infof("exec pod name %s", execPodName)
	// /lotus wallet import --as-default ~/.genesis-sectors/pre-seal-f01000.key ;
	// /lotus-miner init --genesis-miner --actor=f01000 --sector-size=8MiB --pre-sealed-sectors=/root/.genesis-sectors --pre-sealed-metadata=/root/.genesis-sectors/pre-seal-f01000.json --nosync ;
	// /lotus-miner run --nosync ;
	f01000KeyFileName := "/root/.genesis-sectors/pre-seal-f01000.key"
	cmdWalletImport := fmt.Sprintf("/lotus wallet import --as-default %v", f01000KeyFileName)
	importResult, err := k8sEnv.ExecRemoteCmd(ctx, execPodName, "/bin/bash", "-c", cmdWalletImport)
	if err != nil {
		return nil, fmt.Errorf("import key fail %w", err)
	}
	// imported key f3tehwiess4l72p5rfz6rzppx42kcp25clcxhz6mvjghhy6ulqtrom24t5tkarr443lx3e2sso6j7i7d6g6poa successfully!
	re := regexp.MustCompile(`f3[a-z0-9]+`)
	addr, err := address.NewFromString(re.FindString(string(importResult)))
	if err != nil {
		return nil, err
	}

	f01000KeyInfo, err := k8sEnv.ReadSmallFilelInPod(ctx, execPodName, f01000KeyFileName)
	if err != nil {
		return nil, err
	}
	// import keys for wallet
	if err := k8sEnv.ExecRemoteCmdWithStream(ctx, execPodName, true, os.Stdout, nil,
		"/bin/bash", "-c", fmt.Sprintf("/lotus wallet import <<< %v", vrkInfo),
	); err != nil {
		return nil, fmt.Errorf("init genesis miner fail %w", err)
	}
	if err := k8sEnv.ExecRemoteCmdWithStream(ctx, execPodName, true, os.Stdout, nil,
		"/bin/bash", "-c", fmt.Sprintf("/lotus wallet import <<< %v", notaryKey),
	); err != nil {
		return nil, fmt.Errorf("init genesis miner fail %w", err)
	}

	cmdMinerInit := "/lotus-miner init --genesis-miner --actor=f01000 --sector-size=$SECTOR_SIZE --pre-sealed-sectors=/root/.genesis-sectors --pre-sealed-metadata=/root/.genesis-sectors/pre-seal-f01000.json --nosync --confidence 1"
	err = k8sEnv.ExecRemoteCmdWithStream(ctx, execPodName, true, os.Stdout, nil, "/bin/bash", "-c", cmdMinerInit)
	if err != nil {
		return nil, fmt.Errorf("init genesis miner fail %w", err)
	}
	log.Infof("run lotus-miner background")

	cmdMinerRun := "nohup /lotus-miner run --nosync > /root/lotus-miner.log 2>&1 &"
	if err := k8sEnv.ExecRemoteCmdWithStream(ctx, execPodName, false, os.Stdout, nil, "/bin/bash", "-c", cmdMinerRun); err != nil {
		return nil, fmt.Errorf("run lotus-miner fail %w", err)
	}

	var minerInfoErr error
	for i := 0; i < 10; i++ {
		if _, minerInfoErr = k8sEnv.ExecRemoteCmd(ctx, execPodName, "/lotus-miner", "info"); minerInfoErr == nil {
			break
		}
		time.Sleep(time.Second * 3)
	}
	if minerInfoErr != nil {
		return nil, fmt.Errorf("run lotus-miner fail %w", err)
	}
	token, err := k8sEnv.ReadSmallFilelInPod(ctx, execPodName, "/root/.lotus/token")
	if err != nil {
		return nil, err
	}

	libP2pArr, err := k8sEnv.ExecRemoteCmd(ctx, execPodName, "/lotus", "net", "listen")
	if err != nil {
		return nil, fmt.Errorf("exec net listen fail %w", err)
	}

	libP2p := strings.Trim(strings.Split(string(libP2pArr), "\n")[1], " \t\r")
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
	log.Infof("genesis pvc name : %s", claimName)

	//
	vrkAddr, err := address.NewFromString(vrkAddr)
	if err != nil {
		return nil, fmt.Errorf("vrk address %s : %w", vrkAddr, err)
	}
	notaryAddr, err := address.NewFromString(notaryAddr)
	if err != nil {
		return nil, fmt.Errorf("notary address %s : %w", notaryAddr, err)
	}

	return &GenesisReturn{
		Address:        addr,
		GenesisStorage: claimName,
		BootstrapPeers: []string{fmt.Sprintf("/dns/%s/tcp/%s/p2p/%s", svcEndpoint.IP(), port, peer)},
		RPCUrl:         svcEndpoint.ToMultiAddr(),
		RPCToken:       string(token),
		F01000Key:      string(f01000KeyInfo),
		VrkAddr:        vrkAddr,
		NotaryAddr:     notaryAddr,
		CommonDeployParams: env.CommonDeployParams{
			BaseConfig:      incomineCfg.BaseConfig,
			DeployName:      PluginInfo.Name,
			StatefulSetName: statefulSet.GetName(),
			ConfigMapName:   configMap.GetName(),
			SVCName:         svc.GetName(),
			SvcEndpoint:     svcEndpoint,
		},
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

func GetPods(ctx context.Context, k8sEnv *env.K8sEnvDeployer, instanceName string) ([]corev1.Pod, error) {
	return k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("genesis-%s-pod", env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), instanceName)))
}

func InitNotary(ctx context.Context, k8sEnv *env.K8sEnvDeployer, instanceName string) error {
	pods, err := k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("genesis-%s-pod", env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), instanceName)))
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return fmt.Errorf("no genesis pod found")
	}
	// exec pod name
	execPodName := pods[0].Name
	log.Infof("exec pod name %s", execPodName)

	// check balance and add balace for f080 and notary-account  | +10000
	cmdInitNatoary := fmt.Sprintf(`
ACC_NOTARY="%v"
MIN_AMOUNT=1000
ADD_AMOUNT=10000

# add balance
if [ $(/lotus wallet balance f080 | awk '{print int($1)}') -lt $MIN_AMOUNT ]; then
    echo "add $ADD_AMOUNT FIL to f080"
    /lotus send f080 $ADD_AMOUNT &
fi

if [ $(/lotus wallet balance $ACC_NOTARY | awk '{print int($1)}') -lt $MIN_AMOUNT ]; then
    echo "add $ADD_AMOUNT FIL to $ACC_NOTARY"
    /lotus send $ACC_NOTARY $ADD_AMOUNT &
fi

while [ $(/lotus wallet balance f080 | awk '{print int($1)}') -lt $MIN_AMOUNT ] || [ $(/lotus wallet balance $ACC_NOTARY | awk '{print int($1)}') -lt $MIN_AMOUNT ]; do (date && sleep 5); done

# add notary
set_notary_datacap() {
    ACC_NOTARY=$1
    DC_CAP=$2
    FLAG_ADD=0
    # check notary datacap
    while true; do
        RES=$(/lotus filplus check-notary-datacap $ACC_NOTARY 2>/dev/null)
        exit_code=$?
        if [ "$exit_code" -eq 0 ] && [ "$RES" -ge $DC_CAP ]; then
            echo "set-notary-datacap $ACC_NOTARY  $RES ok"
            break
        elif [ "$FLAG_ADD" -eq 0 ]; then
            echo "set notary $ACC_NOTARY datacap $DC_CAP"
            /lotus-shed verifreg add-verifier-from-account f080 $ACC_NOTARY $DC_CAP
            [ $? -ne 0 ] && exit 1
            FLAG_ADD=1
        else
            date && sleep 5
        fi
    done
}
DEFAULT_DC_CAP=100000000000000000
set_notary_datacap $ACC_NOTARY $DEFAULT_DC_CAP
`, notaryAddr)
	if err := k8sEnv.ExecRemoteCmdWithStream(ctx, execPodName, false, os.Stdout, nil, "/bin/bash", "-c", cmdInitNatoary); err != nil {
		return fmt.Errorf("init f080 and notary: %v failed: %w", notaryAddr, err)
	}
	return nil
}

func GrantDatacap(ctx context.Context, k8sEnv *env.K8sEnvDeployer, instanceName string, dcClient string, dcAmount uint64) error {
	pods, err := k8sEnv.GetPodsByLabel(ctx, fmt.Sprintf("genesis-%s-pod", env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), instanceName)))
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return fmt.Errorf("no pod found")
	}
	execPodName := pods[0].Name
	log.Infof("exec pod name %s", execPodName)

	cmdGrant := fmt.Sprintf(`
ACC_NOTARY="%v"
DC_CLIENT="%v"
DC_AMOUNT=%d

set_notary_datacap() {
    ACC_NOTARY=$1
    DC_CAP=$2
    FLAG_ADD=0
    # check notary datacap
    while true; do
        RES=$(/lotus filplus check-notary-datacap $ACC_NOTARY 2>/dev/null)
        exit_code=$?
        if [ "$exit_code" -eq 0 ] && [ "$RES" -ge $DC_CAP ]; then
            echo "set-notary-datacap $ACC_NOTARY  $RES ok"
            break
        elif [ "$FLAG_ADD" -eq 0 ]; then
            echo "set notary $ACC_NOTARY datacap $DC_CAP"
            /lotus-shed verifreg add-verifier-from-account f080 $ACC_NOTARY $DC_CAP
            [ $? -ne 0 ] && exit 1
            FLAG_ADD=1
        else
            date && sleep 5
        fi
    done
}

# grant datacap
FLAG_ADD=0
ORIGIN_AMOUNT=0
RES=$(/lotus filplus check-client-datacap $DC_CLIENT) && ORIGIN_AMOUNT=$RES
TARGET_AMOUNT=$((ORIGIN_AMOUNT + DC_AMOUNT))
while true; do
    RES=$(/lotus filplus check-client-datacap $DC_CLIENT)
    exit_code=$?
    echo "check-client-datacap $DC_CLIENT | $RES"
    if [ "$exit_code" -eq 0 ] && [ "$RES" -eq $TARGET_AMOUNT ]; then
        break
    elif [ "$FLAG_ADD" -eq 0 ]; then
        set_notary_datacap $ACC_NOTARY $DC_AMOUNT
        echo "Adding verifier client $DC_CLIENT datacap $DC_AMOUNT"
        /lotus filplus grant-datacap --from $ACC_NOTARY $DC_CLIENT $DC_AMOUNT
        [ $? -ne 0 ] && exit 1
        FLAG_ADD=1
    else
        date && sleep 5
    fi
done`, notaryAddr, dcClient, dcAmount)
	if err := k8sEnv.ExecRemoteCmdWithStream(ctx, execPodName, false, os.Stdout, nil, "/bin/bash", "-c", cmdGrant); err != nil {
		return fmt.Errorf("grant datacap fail %w", err)
	}

	return nil
}
