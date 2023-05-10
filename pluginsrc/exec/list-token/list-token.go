package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "generate_token",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "generate admin token",
}

type TestCaseParams struct {
	fx.In
	Params struct {
		skip  string `json:"skip"`
		limit string `json:"limit"`
	} `optional:"true"`

	K8sEnv    *env.K8sEnvDeployer `json:"-"`
	VenusAuth env.IDeployer       `json:"-" svcname:"VenusAuth"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	endpoint := params.VenusAuth.SvcEndpoint()
	if env.Debug {
		pods, err := params.VenusAuth.Pods(ctx)
		if err != nil {
			return nil, err
		}

		svc, err := params.VenusAuth.Svc(ctx)
		if err != nil {
			return nil, err
		}
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return nil, err
		}
	}

	adminTokenV, err := params.VenusAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	authAPIClient, err := jwtclient.NewAuthClient(endpoint.ToHttp(), adminTokenV.(string))
	if err != nil {
		return nil, err
	}

	skip, err := strconv.ParseInt(params.Params.skip, 10, 64)
	limit, err := strconv.ParseInt(params.Params.limit, 10, 64)
	_, err = authAPIClient.Tokens(ctx, skip, limit)
	if err != nil {
		return nil, err
	}

	adminToken, err := authAPIClient.GenerateToken(ctx, "admin", "admin", "")
	if err != nil {
		return nil, err
	}
	fmt.Println(adminToken)
	return env.NewSimpleExec(), nil
}
