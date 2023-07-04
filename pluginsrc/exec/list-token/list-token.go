package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"github.com/ipfs-force-community/sophon-auth/jwtclient"
	"go.uber.org/fx"
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
	fx.In
	Params struct {
		Skip  string `json:"skip"`
		Limit string `json:"limit"`
	} `optional:"true"`

	K8sEnv     *env.K8sEnvDeployer `json:"-"`
	SophonAuth env.IDeployer       `json:"-" svcname:"SophonAuth"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	endpoint, err := params.SophonAuth.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	adminTokenV, err := params.SophonAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	authAPIClient, err := jwtclient.NewAuthClient(endpoint.ToHTTP(), adminTokenV.MustString())
	if err != nil {
		return nil, err
	}

	skip, err := strconv.ParseInt(params.Params.Skip, 10, 64)
	if err != nil {
		return nil, err
	}

	limit, err := strconv.ParseInt(params.Params.Limit, 10, 64)
	if err != nil {
		return nil, err
	}
	tokenList, err := authAPIClient.Tokens(ctx, skip, limit)
	if err != nil {
		return nil, err
	}
	for _, token := range tokenList {
		fmt.Println(token)
	}
	return env.NewSimpleExec(), nil
}
