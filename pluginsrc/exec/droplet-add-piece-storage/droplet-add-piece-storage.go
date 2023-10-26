package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletmarket "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/pvc"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("droplet-add-piece-storage")

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
	err := dropletmarket.AddPieceStoragge(ctx, k8sEnv, params.DropletMarket, params.PieceStore.Name, params.MountPath)
	if err != nil {
		return err
	}

	pods, err := dropletmarket.GetPods(ctx, k8sEnv, params.DropletMarket.InstanceName)
	if err != nil {
		return err
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
