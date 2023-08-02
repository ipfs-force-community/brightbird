package main

import (
	"context"

	chain "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "auth_request_venus",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "auth request venus",
}

type TestCaseParams struct {
	Auth  sophonauth.SophonAuthDeployReturn `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Venus venus.VenusDeployReturn           `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	err := checkPermission(ctx, params.Auth.AdminToken, params)
	if err != nil {
		return err
	}
	return nil
}

func checkPermission(ctx context.Context, token string, params TestCaseParams) error {
	chainRPC, closer, err := chain.DialFullNodeRPC(ctx, params.Venus.SvcEndpoint.ToMultiAddr(), token, nil)
	if err != nil {
		return err
	}
	defer closer()

	_, err = chainRPC.ChainHead(ctx)
	if err != nil {
		return err
	}

	return nil
}
