package main

import (
	"context"

	chain "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/hunjixin/brightbird/pluginsrc/deploy/venus"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
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
	SophonAuth sophonauth.SophonAuthDeployReturn `json:"SophonAuth" description:"sophon auth return"`
	Venus      venus.VenusDeployReturn           `json:"Venus" description:"Venus return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	err := checkPermission(ctx, params.SophonAuth.AdminToken, params)
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
