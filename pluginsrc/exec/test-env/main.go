package main

import (
	"context"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("test-env")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "test-env",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "test global params",
}

type TestCaseParams struct {
	Global *env.GlobalParams `jsonschema:"-" container:"type"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	stringVal, err := params.Global.GetProperty("aaa")
	plugin.Assert.NoError(err)
	log.Infof("get strignVal %s", stringVal)

	numberVal, err := params.Global.GetNumberProperty("bbb")
	plugin.Assert.NoError(err)
	log.Infof("get strignVal %v", numberVal)

	arr := []string{}
	err = params.Global.GetJSONProperty("ccc", &arr)
	plugin.Assert.NoError(err)
	log.Infof("get arraryval %v", arr)
	return nil
}
