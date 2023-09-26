package pvc

import (
	"context"
	"embed"
	"fmt"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

type Config struct {
	env.BaseConfig
}

type RenderParams struct {
	Name      string
	NameSpace string
	UniqueId  string
}

var PluginInfo = types.PluginInfo{
	Name:               "pvc",
	Version:            version.Version(),
	PluginType:         types.Deploy,
	DeployPluginParams: types.DeployPluginParams{},
	Description:        "",
}

type PvcReturn struct { //nolint
	Name string `json:"name" jsonschema:"name" title:"PVC UniqueName" require:"true" description:"pvc's unique name"`
}

//go:embed pvc.yaml
var f embed.FS

func DeployFromConfig(ctx context.Context, k8sEnv *env.K8sEnvDeployer, cfg Config) (*PvcReturn, error) {
	pvcYamlData, err := f.Open("pvc.yaml")
	if err != nil {
		return nil, err
	}
	pvc, err := k8sEnv.CreatePvc(ctx, pvcYamlData, RenderParams{
		Name:      cfg.InstanceName,
		NameSpace: k8sEnv.NameSpace(),
		UniqueId:  env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), cfg.InstanceName),
	})
	if err != nil {
		return nil, fmt.Errorf("create pvc fail %w", err)
	}
	return &PvcReturn{
		Name: pvc.Name,
	}, nil
}
