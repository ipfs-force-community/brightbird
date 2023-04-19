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
	Name:        "diff_permissions_token",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "diff permissions token",
}

type TestCaseParams struct {
	fx.In
	Params struct {
		Permission string `json:"permission"`
	} `optional:"true"`
	AdminToken types.AdminToken
	K8sEnv     *env.K8sEnvDeployer    `json:"-"`
	VenusAuth  env.IVenusAuthDeployer `json:"-"`
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
	authAPIClient, err := jwtclient.NewAuthClient(endpoint.ToHttp(), string(params.AdminToken))
	if err != nil {
		return err
	}

	_, err = authAPIClient.CreateUser(ctx, &auth.CreateUserRequest{
		Name:    params.Params.Permission,
		Comment: utils.StringPtr("comment " + params.Params.Permission),
		State:   0,
	})
	if err != nil {
		return err
	}

	token, err := authAPIClient.GenerateToken(ctx, params.Params.Permission, params.Params.Permission, "")
	if err != nil {
		return err
	}
	fmt.Println(token)

	switch params.Params.Permission {
	case "admin":
		fmt.Println("num is 1")
	case "sign":
		fmt.Println("num is 2")
	case "write":
		fmt.Println("num is 3")
	case "read":
		fmt.Println("num is 4")
	default:
		fmt.Println("Incorrect permission type")
	}
	return nil
}
