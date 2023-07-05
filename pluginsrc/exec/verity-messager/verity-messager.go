package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	sophonmessager "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-messager"
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
	SophonAuth     sophonauth.SophonAuthDeployReturn   `json:"SophonAuth"`
	SophonMessager sophonmessager.SophonMessagerReturn `json:"SophonMessager"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	client, closer, err := messager.DialIMessagerRPC(ctx, params.SophonMessager.SvcEndpoint.ToMultiAddr(), params.SophonAuth.AdminToken, nil)
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
