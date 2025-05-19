package main

import (
	"context"
	"fmt"
	"path/filepath"

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
	Name:        "client-create-file",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "create file use droplet-client",
}

type TestCaseParams struct {
	DropletClient dropletclient.DropletClientDeployReturn `json:"DropClient" jsonschema:"DropClient" title:"DropletClient" description:"droplet client return"`
	Droplet       dropletmarket.DropletMarketDeployReturn `json:"Droplet" jsonschema:"Droplet" title:"Droplet" description:"droplet return"`
	PieceStore    pvc.PvcReturn                           `json:"PieceStore" jsonschema:"PieceStore" title:"PieceStore" require:"true" description:"piece storage"`

	// FileSize需要在（.droplet storage ask set）设置的 minPrice和maxPrice之间）
	FileSize string `json:"FileSize" jsonschema:"FileSize" title:"FileSize" default:"512" require:"true" description:"File size in bytes (b=512, kB=1000, K=1024, MB=kB*kB, M=K*K, GB=kB*kB*kB, G=K*K*K)"`
}

type ClientCreateFileReturn struct {
	File    string
	CarFile string
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*ClientCreateFileReturn, error) {
	pods, err := dropletclient.GetPods(ctx, k8sEnv, params.DropletClient.InstanceName)
	if err != nil {
		return nil, err
	}
	if len(pods) == 0 {
		return nil, fmt.Errorf("not found pod for %v", params.DropletClient.InstanceName)
	}
	execPodName := pods[0].GetName()

	// add pvc for droplet-client
	mountPath := filepath.Join("/carfile/", params.PieceStore.Name)
	if err := dropletclient.AddPieceStoragge(ctx, k8sEnv, params.DropletClient, params.PieceStore.Name, &mountPath); err != nil {
		return nil, err
	}

	filePath := filepath.Join(mountPath, "/file.txt")
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

	log.With(
		"filePath", filePath,
		"carFilePath", carFile,
	).Info("res")

	return &ClientCreateFileReturn{
		File:    filePath,
		CarFile: carFile,
	}, nil
}
