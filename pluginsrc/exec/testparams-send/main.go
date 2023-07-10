package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/pluginsrc/exec/testparams-send/testparamssend"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
)

func main() {
	plugin.SetupPluginFromStdin(PluginInfo, Exec)
}

var PluginInfo = types.PluginInfo{
	Name:        "test-params-send",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "",
}

type DepParams struct {
	testparamssend.Config
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*testparamssend.TestDeployReturn, error) {
	return &testparamssend.TestDeployReturn{
		Config: testparamssend.Config{
			NumberTest:    depParams.NumberTest,
			IntegerTest:   depParams.IntegerTest,
			StringTest:    depParams.StringTest,
			StringArrTest: depParams.StringArrTest,
			EmbedStruct: testparamssend.EmbedStruct{
				NumberTest:    depParams.NumberTest,
				IntegerTest:   depParams.IntegerTest,
				StringTest:    depParams.StringTest,
				StringArrTest: depParams.StringArrTest,
			},
			EmbedArrayStruct: []testparamssend.EmbedStruct{
				{
					NumberTest:    depParams.NumberTest,
					IntegerTest:   depParams.IntegerTest,
					StringTest:    depParams.StringTest,
					StringArrTest: depParams.StringArrTest,
				},
			},
		},
	}, nil
}
