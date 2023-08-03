package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	testparamssend "github.com/ipfs-force-community/brightbird/pluginsrc/exec/testparams-send/testparamssend"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

func main() {
	plugin.SetupPluginFromStdin(PluginInfo, Exec)
}

var PluginInfo = types.PluginInfo{
	Name:        "test-params-receive",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "",
}

type DepParams struct {
	testparamssend.Config
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) error {
	data, err := json.MarshalIndent(depParams.Config, "\t", " ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
