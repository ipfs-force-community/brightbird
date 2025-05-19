package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"

	dropletmarket "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/pvc"
)

var log = logging.Logger(Info.Name)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "droplet-add-piece-storage",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "add a new piece storage in droplet",
}

type TestCaseParams struct {
	DropletMarket dropletmarket.DropletMarketDeployReturn `json:"DropletMarket" jsonschema:"DropletMarket" title:"DropletMarket" description:"droplet market return"`
	PieceStore    pvc.PvcReturn                           `json:"PieceStore" jsonschema:"PieceStore" title:"PieceStore" require:"true" description:"piece storage"`
	MountPath     string                                  `json:"MountPath" jsonschema:"MountPath" title:"MountPath" require:"true" description:"/piece/"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	err := dropletmarket.AddPieceStoragge(ctx, k8sEnv, params.DropletMarket, params.PieceStore.Name, filepath.Join(params.MountPath, params.PieceStore.Name))
	if err != nil {
		return err
	}

	pods, err := dropletmarket.GetPods(ctx, k8sEnv, params.DropletMarket.InstanceName)
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return fmt.Errorf("not found pod for %v", params.DropletMarket.InstanceName)
	}

	pieceList, err := k8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), "/bin/bash", "-c", "./droplet piece-storage list")
	if err != nil {
		return err
	}

	log.Infof("piece storage list %s", string(pieceList))
	if strings.Contains(string(pieceList), params.PieceStore.Name) {
		return nil
	}
	return fmt.Errorf("check new storage fail")
}
