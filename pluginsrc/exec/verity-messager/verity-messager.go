package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	sophonmessager "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-messager"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
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
	Auth     sophonauth.SophonAuthDeployReturn   `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Messager sophonmessager.SophonMessagerReturn `json:"SophonMessager"  jsonschema:"SophonMessager"  title:"Sophon Messager" require:"true" description:"messager return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	client, closer, err := messager.DialIMessagerRPC(ctx, params.Messager.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
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
