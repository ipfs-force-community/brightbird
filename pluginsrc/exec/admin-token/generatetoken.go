package main

import (
	"context"
	"fmt"

	"github.com/hunjixin/brightbird/types"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "generate_token",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "generate admin token",
}

type TestCaseParams struct {
	fx.In

	K8sEnv    *env.K8sEnvDeployer `json:"-"`
	VenusAuth env.IDeployer       `json:"-" svcname:"VenusAuth"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	venusAuthPods, err := params.VenusAuth.Pods(ctx)
	if err != nil {
		return nil, err
	}

	svc, err := params.VenusAuth.Svc(ctx)
	if err != nil {
		return nil, err
	}

	endpoint, err := params.VenusAuth.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, venusAuthPods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return nil, err
		}
	}

	adminToken, err := params.VenusAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}
	authAPIClient, err := jwtclient.NewAuthClient(endpoint.ToHttp(), adminToken.MustString())
	if err != nil {
		return nil, err
	}

	_, err = authAPIClient.CreateUser(ctx, &auth.CreateUserRequest{
		Name:    "admin",
		Comment: utils.StringPtr("comment admin"),
		State:   0,
	})
	if err != nil {
		return nil, err
	}

	adminTokenv, err := authAPIClient.GenerateToken(ctx, "admin", "admin", "")
	if err != nil {
		return nil, err
	}
	fmt.Println(adminTokenv)
	return env.NewSimpleExec(), nil
}
