package main

import (
	"context"
	"fmt"
	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
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
	K8sEnv    *env.K8sEnvDeployer    `json:"-"`
	VenusAuth env.IVenusAuthDeployer `json:"-"`
}

func Exec(ctx context.Context, params TestCaseParams) error {
	endpoint := params.VenusAuth.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, params.VenusAuth.Pods()[0].GetName(), int(params.VenusAuth.Svc().Spec.Ports[0].Port))
		if err != nil {
			return err
		}
	}
	authAPIClient, err := jwtclient.NewAuthClient(endpoint.ToHttp())
	if err != nil {
		return err
	}

	_, err = authAPIClient.CreateUser(ctx, &auth.CreateUserRequest{
		Name:    "admin",
		Comment: utils.StringPtr("comment admin"),
		State:   0,
	})
	if err != nil {
		return err
	}

	adminToken, err := authAPIClient.GenerateToken(ctx, "admin", "admin", "")
	if err != nil {
		return err
	}
	fmt.Println(adminToken)
	return nil
}
