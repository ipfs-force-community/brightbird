package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"github.com/ipfs-force-community/sophon-auth/jwtclient"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "list_token",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "list token",
}

type TestCaseParams struct {
	Skip       string                            `json:"skip"`
	Limit      string                            `json:"limit"`
	SophonAuth sophonauth.SophonAuthDeployReturn `json:"SophonAuth"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	authAPIClient, err := jwtclient.NewAuthClient(params.SophonAuth.SvcEndpoint.ToHTTP(), params.SophonAuth.AdminToken)
	if err != nil {
		return err
	}

	skip, err := strconv.ParseInt(params.Skip, 10, 64)
	if err != nil {
		return err
	}

	limit, err := strconv.ParseInt(params.Limit, 10, 64)
	if err != nil {
		return err
	}
	tokenList, err := authAPIClient.Tokens(ctx, skip, limit)
	if err != nil {
		return err
	}
	for _, token := range tokenList {
		fmt.Println(token)
	}
	return nil
}
