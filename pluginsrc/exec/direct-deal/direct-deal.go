package main

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/filecoin-project/go-address"
	clientapi "github.com/filecoin-project/venus/venus-shared/api/market/client"
	"github.com/filecoin-project/venus/venus-shared/types/market/client"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"

	dropletclient "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-client"
	dropletmarket "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/pvc"
)

var log = logging.Logger(Info.Name)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "direct-deal",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "droplet new direct-deal and import to market",
}

type TestCaseParams struct {
	DropletClient dropletclient.DropletClientDeployReturn `json:"DropClient" jsonschema:"DropClient" title:"DropletClient" description:"droplet client return"`
	Droplet       dropletmarket.DropletMarketDeployReturn `json:"Droplet" jsonschema:"Droplet" title:"Droplet" description:"droplet return"`

	PieceStore pvc.PvcReturn `json:"PieceStore" jsonschema:"PieceStore" title:"PieceStore" require:"true" description:"piece storage"`

	// FileSize需要在（.droplet storage ask set）设置的 minPrice和maxPrice之间）
	FileSize string `json:"FileSize" jsonschema:"FileSize" title:"FileSize" default:"512" require:"true" description:"File size in bytes (b=512, kB=1000, K=1024, MB=kB*kB, M=K*K, GB=kB*kB*kB, G=K*K*K)"`

	From         address.Address `json:"From"  jsonschema:"From"  title:"From" require:"false" description:"From"`
	MinerAddress address.Address `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true"`
	Duration     int64           `json:"Duration"  jsonschema:"Duration"  title:"Duration" default:"518400" require:"false" description:"Set the price of the ask for retrievals"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {

	// init pvc and file
	pieceCompInfo, err := initPvcAndFile(ctx, k8sEnv, params)
	if err != nil {
		return err
	}
	log.With(
		"Root", pieceCompInfo.Root,
		"PieceSize", pieceCompInfo.Size,
	).Info("piece info")

	// direct-deal allocate
	allocationFile, err := CreateDirectDeal(ctx, k8sEnv, params, pieceCompInfo)
	if err != nil {
		return err
	}
	// import deal to market
	if err := ImportDirectDeal(ctx, k8sEnv, params, allocationFile); err != nil {
		return err
	}
	return nil
}
func initPvcAndFile(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*client.CommPRet, error) {
	api, closer, err := clientapi.DialIMarketClientRPC(ctx, params.DropletClient.SvcEndpoint.ToMultiAddr(), params.DropletClient.ClientToken, nil)
	if err != nil {
		return nil, fmt.Errorf("dial market client rpc failed")
	}
	defer closer()

	pods, err := dropletclient.GetPods(ctx, k8sEnv, params.DropletClient.InstanceName)
	if err != nil {
		return nil, err
	}
	if len(pods) == 0 {
		return nil, fmt.Errorf("not found pod for %v", params.DropletClient.InstanceName)
	}
	execPodName := pods[0].GetName()

	// add pvc for droplet-client
	mountPath := filepath.Join("/piece/", params.PieceStore.Name)
	if err := dropletclient.AddPieceStoragge(ctx, k8sEnv, params.DropletClient, params.PieceStore.Name, &mountPath); err != nil {
		return nil, err
	}

	filePath := filepath.Join(mountPath, "file.txt")
	if _, err := k8sEnv.ExecRemoteCmd(ctx, execPodName,
		"/bin/sh", "-c",
		fmt.Sprintf("dd if=/dev/urandom of=%v bs=%v count=1", filePath, params.FileSize)); err != nil {
		return nil, err
	}
	carFile := filePath + ".car"
	if _, err := k8sEnv.ExecRemoteCmd(ctx, execPodName,
		"/bin/sh", "-c",
		fmt.Sprintf("./droplet-client data generate-car %v %v", filePath, carFile)); err != nil {
		return nil, err
	}

	// mount pvc in droplet
	if err := dropletmarket.AddPieceStoragge(ctx, k8sEnv, params.Droplet, params.PieceStore.Name, mountPath); err != nil {
		return nil, err
	}

	pieceCompInfo, err := api.ClientCalcCommP(ctx, carFile)
	if err != nil {
		return nil, err
	}
	pieceName := filepath.Join(mountPath, pieceCompInfo.Root.String())
	if _, err := k8sEnv.ExecRemoteCmd(ctx, execPodName,
		"/bin/sh", "-c",
		fmt.Sprintf("mv %v %v", carFile, pieceName)); err != nil {
		return nil, err
	}

	log.With(
		"filePath", filePath,
		"pieceFile", pieceName,
	).Info("res")

	return pieceCompInfo, nil
}

func CreateDirectDeal(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams, initData *client.CommPRet) (string, error) {
	instanceName := params.DropletClient.InstanceName
	pods, err := dropletclient.GetPods(ctx, k8sEnv, instanceName)
	if err != nil {
		return "", err
	}
	if len(pods) == 0 {
		return "", fmt.Errorf("not found pod for %v", instanceName)
	}
	execPodName := pods[0].GetName()
	pvcMountPath, err := dropletclient.GetPvcMountPath(ctx, k8sEnv, params.DropletClient, params.PieceStore.Name)
	if err != nil {
		return "", err
	}

	log.Infof("exec in pod : %v", execPodName)
	outAllocationFile := fmt.Sprintf("allocation-%v.csv", time.Now().Unix())
	ddoCmd := fmt.Sprintf(`./droplet-client direct-deal allocate \
--miner %v --wallet %v \
--piece-info %v=%v \
--tmin %v --output-allocation-to-file %v`,
		params.MinerAddress.String(), params.From,
		initData.Root.String(), initData.Size,
		params.Duration, filepath.Join(pvcMountPath, outAllocationFile))
	if res, err := k8sEnv.ExecRemoteCmd(ctx, execPodName, "/bin/sh", "-c", ddoCmd); err != nil {
		return "", fmt.Errorf("exec [%v] failed: %v", ddoCmd, err)
	} else {
		log.Infof("\n%v", string(res))
	}

	return outAllocationFile, nil
}
func ImportDirectDeal(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams, allocationFile string) error {
	instanceName := params.Droplet.InstanceName
	pods, err := dropletmarket.GetPods(ctx, k8sEnv, instanceName)
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return fmt.Errorf("not found pod for %v", instanceName)
	}
	execPodName := pods[0].GetName()
	pvcMountPath, err := dropletmarket.GetPvcMountPath(ctx, k8sEnv, params.Droplet, params.PieceStore.Name)
	if err != nil {
		return err
	}

	log.Infof("exec in pod : %v", execPodName)
	importDealsCmd := fmt.Sprintf(`./droplet storage direct-deal import-deals --allocation-file %v --car-dir %v`,
		filepath.Join(pvcMountPath, allocationFile), pvcMountPath)
	maxRetry := 3
	for i := 0; i < maxRetry; i++ {
		if res, err := k8sEnv.ExecRemoteCmd(ctx, execPodName, "/bin/sh", "-c", importDealsCmd); err != nil {
			if i+1 >= maxRetry {
				return fmt.Errorf("exec [%v] failed: %v", importDealsCmd, err)
			}
			log.Warnf("exec import-deals failed: %v, retry %d/%d", err, i+1, maxRetry)
			time.Sleep(30 * time.Second)
		} else {
			log.Infof("\n%v", string(res))
			break
		}
	}
	return nil
}

// // script for wait direct deal state
// out_tmp_file="/tmp/direct-deal-list.tmp"
// LAST_RES=""
// while true; do
//     for i in $(seq 1 6); do
//         TMP_RES=$(/app/droplet storage direct-deal list --state $i)
//         [ "$(echo "$TMP_RES" | wc -l)" -gt 1 ] && echo "$TMP_RES" >>"$out_tmp_file"
//     done

//     if [ -f "$out_tmp_file" ]; then
//         RES=$(sort -r "$out_tmp_file" | uniq)
//         [ "$LAST_RES" != "$RES" ] && [ "$(echo "$RES" | wc -l)" -gt 1 ] &&
//             echo "$RES" && LAST_RES="$RES" &&
//             [ "$(echo "$RES" | grep -E "DealActive|DealExpired|DealSlashed|DealError" | wc -l)" -gt 0 ] &&
//             echo "----- direct deal end -----" &&
//             exit
//     fi
//     echo "---- $(date) ----" && sleep 60
//     rm -rf $out_tmp_file
// done
