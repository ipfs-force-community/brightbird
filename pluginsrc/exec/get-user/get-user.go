package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "get_user",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "get user name",
}

type TestCaseParams struct {
	fx.In
	Params struct {
		UserName string `json:"userName"`
	} `optional:"true"`

	K8sEnv    *env.K8sEnvDeployer `json:"-"`
	VenusAuth env.IDeployer       `json:"-" svcname:"VenusAuth"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	endpoint := params.VenusAuth.SvcEndpoint()
	if env.Debug {
		venusAuthPods, err := params.VenusAuth.Pods(ctx)
		if err != nil {
			return nil, err
		}

		svc, err := params.VenusAuth.Svc(ctx)
		if err != nil {
			return nil, err
		}
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, venusAuthPods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return nil, err
		}
	}
	adminToken, err := params.VenusAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	authAPIClient, err := jwtclient.NewAuthClient(endpoint.ToHttp(), adminToken.(string))
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
