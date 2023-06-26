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
	Name:        "create_token",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "create token",
}

type TestCaseParams struct {
	fx.In
	Params struct {
		Perm  string `json:"perm" description:"[OPTIONS] custom string in JWT payload"`
		Extra string `json:"extra" description:"[OPTIONS] permission for API auth (read, write, sign, admin)"`
	} `optional:"true"`

	K8sEnv     *env.K8sEnvDeployer `json:"-"`
	UserName   env.IExec           `json:"-" svcname:"UserName" description:"[Exec]create-user"`
	SophonAuth env.IDeployer       `json:"-" svcname:"SophonAuth" description:"[Deploy]venus-auth"`
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

	userName, err := params.UserName.Param("UserName")
	if err != nil {
		return nil, err
	}
	if len(userName.MustString()) == 0 {
		return nil, fmt.Errorf("specific user name")
	}

	token, err := authAPIClient.GenerateToken(ctx, userName.MustString(), params.Params.Perm, params.Params.Extra)
	if err != nil {
		return nil, err
	}

	return env.NewSimpleExec().Add("Token", env.ParamsFromVal(token)), nil
}
