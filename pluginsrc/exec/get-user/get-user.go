package main

import (
	"context"
	"fmt"

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
	Name:        "get_user",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "get user name",
}

type TestCaseParams struct {
	fx.In
	Params struct {
		UserName string `json:"userName"`
	} `optional:"true"`

	K8sEnv     *env.K8sEnvDeployer `json:"-"`
	SophonAuth env.IDeployer       `json:"-" svcname:"SophonAuth"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	endpoint, err := params.SophonAuth.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	adminToken, err := params.SophonAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	authAPIClient, err := jwtclient.NewAuthClient(endpoint.ToHTTP(), adminToken.MustString())
	if err != nil {
		return nil, err
	}

	user, err := authAPIClient.GetUser(ctx, params.Params.UserName)
	if err != nil {
		return nil, err
	}

	fmt.Println(user.Name)
	return env.NewSimpleExec(), nil
}
