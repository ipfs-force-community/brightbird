package main

import (
	"context"
	"fmt"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"github.com/ipfs-force-community/sophon-auth/auth"
	"github.com/ipfs-force-community/sophon-auth/core"
	"github.com/ipfs-force-community/sophon-auth/jwtclient"
	"go.uber.org/fx"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "create_user",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "create user token",
}

type TestCaseParams struct {
	fx.In
	Params struct {
		UserName string `json:"userName"`
		Comment  string `json:"comment"`
	} `optional:"true"`

	K8sEnv     *env.K8sEnvDeployer `json:"-"`
	SophonAuth env.IDeployer       `json:"-" svcname:"SophonAuth"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	endpoint, err := params.SophonAuth.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	if env.Debug {
		venusAuthPods, err := params.SophonAuth.Pods(ctx)
		if err != nil {
			return nil, err
		}

		svc, err := params.SophonAuth.Svc(ctx)
		if err != nil {
			return nil, err
		}
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, venusAuthPods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return nil, err
		}
	}
	adminToken, err := params.SophonAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	authAPIClient, err := jwtclient.NewAuthClient(endpoint.ToHTTP(), adminToken.MustString())
	if err != nil {
		return nil, err
	}

	if len(params.Params.UserName) == 0 {
		return nil, fmt.Errorf("username cant be empty")
	}

	user, err := authAPIClient.CreateUser(ctx, &auth.CreateUserRequest{
		Name:    params.Params.UserName,
		Comment: &params.Params.Comment,
		State:   core.UserStateEnabled,
	})
	if err != nil {
		return nil, err
	}

	fmt.Println(user.Name)
	return env.NewSimpleExec().Add("UserName", env.ParamsFromVal(params.Params.UserName)), nil
}
