package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/jwtclient"
	chain "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "auth_request_venus",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "auth request venus",
}

type TestCaseParams struct {
	fx.In
	Params struct {
		//Permission string `json:"permission"`
	} `optional:"true"`
	AdminToken types.AdminToken
	K8sEnv     *env.K8sEnvDeployer    `json:"-"`
	VenusAuth  env.IVenusAuthDeployer `json:"-"`
	Venus      env.IVenusDeployer     `json:"-"`
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
		Name:    "admin",
		Comment: utils.StringPtr("comment admin"),
		State:   0,
	})
	if err != nil {
		return err
	}

	token, err := authAPIClient.GenerateToken(ctx, "admin", "admin", "")
	if err != nil {
		return err
	}
	fmt.Println(token)

	err = checkPermission(ctx, token, params)
	if err != nil {
		return err
	}
	return nil
}

func checkPermission(ctx context.Context, token string, params TestCaseParams) error {
	endpoint := params.Venus.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, params.Venus.Pods()[0].GetName(), int(params.Venus.Svc().Spec.Ports[0].Port))
		if err != nil {
			return err
		}
	}
	chainRpc, closer, err := chain.DialFullNodeRPC(ctx, endpoint.ToMultiAddr(), token, nil)
	if err != nil {
		return err
	}
	defer closer()

	_, err = chainRpc.ChainHead(ctx)
	if err != nil {
		return err
	}

	return nil
}
