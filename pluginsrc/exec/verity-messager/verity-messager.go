package main

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "verity_message",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "verity message if normal",
}

type TestCaseParams struct {
	fx.In
	K8sEnv         *env.K8sEnvDeployer `json:"-"`
	SophonMessager env.IDeployer       `json:"-" svcname:"SophonMessager"`
	SophonAuth     env.IDeployer       `json:"-" svcname:"SophonAuth"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {

	adminTokenV, err := params.SophonAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	err = CreateMessage(ctx, params, adminTokenV.MustString())
	if err != nil {
		fmt.Printf("create message rpc failed: %v\n", err)
		return nil, err
	}

	return env.NewSimpleExec(), nil
}

func CreateMessage(ctx context.Context, params TestCaseParams, authToken string) error {
	endpoint, err := params.SophonMessager.SvcEndpoint()
	if err != nil {
		return err
	}

	client, closer, err := messager.DialIMessagerRPC(ctx, endpoint.ToHTTP(), authToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	messageVersion, err := client.Version(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Message: %v\n", messageVersion)

	return nil
}
