package main

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletclient "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-client"
	dropletmarket "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/pvc"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("client-create-file")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "client-create-file",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "在client创建文件",
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
	mountPath := "/carfile/"
	filePath := mountPath + params.PieceStore.Name + "/file.txt"
	carFile := filePath + ".car"
	log.Debugln("carFilePath:", carFile)

	tmpl, err := template.New("command").Parse("dd if=/dev/urandom of={{.FilePath}} bs={{.BlockSize}} count=1")
	if err != nil {
		return nil, fmt.Errorf("parase template: %v", err)
	}

	data := map[string]interface{}{
		"FilePath":  filePath,
		"BlockSize": params.FileSize,
	}

	var createFileCmd bytes.Buffer
	err = tmpl.Execute(&createFileCmd, data)
	if err != nil {
		panic(err)
	}

	err = dropletclient.AddPieceStoragge(ctx, k8sEnv, params.DropletClient, params.PieceStore.Name, mountPath)
	if err != nil {
		return nil, err
	}

	pods, err := dropletclient.GetPods(ctx, k8sEnv, params.DropletClient.InstanceName)
	if err != nil {
		return nil, err
	}

	_, err = k8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), "/bin/sh", "-c", createFileCmd.String())
	if err != nil {
		return nil, err
	}

	// mount pvc in droplet
	err = dropletmarket.AddPieceStoragge(ctx, k8sEnv, params.Droplet, params.PieceStore.Name, mountPath)
	if err != nil {
		return nil, err
	}

	return &ClientCreateFileReturn{
		File:    filePath,
		CarFile: carFile,
	}, nil
}
