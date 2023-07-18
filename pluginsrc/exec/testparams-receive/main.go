package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	testparamssend "github.com/hunjixin/brightbird/pluginsrc/exec/testparams-send/testparamssend"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
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
