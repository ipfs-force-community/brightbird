package main

import (
	"context"
	"fmt"
	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "verity_gateway",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "verity gateway if normal",
}

type TestCaseParams struct {
	fx.In
	AdminToken   types.AdminToken
	K8sEnv       *env.K8sEnvDeployer       `json:"-"`
	VenusAuth    env.IVenusAuthDeployer    `json:"-"`
	VenusMessage env.IVenusMessageDeployer `json:"-"`
}

func Exec(ctx context.Context, params TestCaseParams) error {
	authToken, err := CreateAuthToken(ctx, params)
	if err != nil {
		fmt.Printf("create auth token failed: %v\n", err)
		return err
	}

	err = CreateMessage(ctx, params, authToken)
	if err != nil {
		fmt.Printf("create message rpc failed: %v\n", err)
		return err
	}

	return nil

}

func CreateAuthToken(ctx context.Context, params TestCaseParams) (adminToken string, err error) {
	endpoint := params.VenusAuth.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, params.VenusAuth.Pods()[0].GetName(), int(params.VenusAuth.Svc().Spec.Ports[0].Port))
		if err != nil {
			return "", err
		}
	}

	authAPIClient, err := jwtclient.NewAuthClient(endpoint.ToHttp(), string(params.AdminToken))
	if err != nil {
		return "", err
	}
	_, err = authAPIClient.CreateUser(ctx, &auth.CreateUserRequest{
		Name:    "admin",
		Comment: utils.StringPtr("comment admin"),
		State:   0,
	})
	if err != nil {
		return "", err
	}

	adminToken, err = authAPIClient.GenerateToken(ctx, "admin", "admin", "")
	if err != nil {
		return "", err
	}

	return adminToken, nil
}

func CreateMessage(ctx context.Context, params TestCaseParams, authToken string) error {
	endpoint := params.VenusMessage.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, params.VenusMessage.Pods()[0].GetName(), int(params.VenusMessage.Svc().Spec.Ports[0].Port))
		if err != nil {
			return err
		}
	}

	client, closer, err := messager.DialIMessagerRPC(ctx, endpoint.ToHttp(), authToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	messageVersion, err := client.Version(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Message: %v\n", messageVersion)

	return nil
}
