package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"

	testparamssend "github.com/ipfs-force-community/brightbird/pluginsrc/exec/testparams-send/testparamssend"
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
