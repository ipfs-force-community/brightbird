package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/types"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "verity_message",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "verity message if normal",
}

type TestCaseParams struct {
	fx.In
	K8sEnv       *env.K8sEnvDeployer `json:"-"`
	VenusAuth    env.IDeployer       `json:"-" svcname:"VenusAuth"`
	VenusMessage env.IDeployer       `json:"-" svcname:"VenusMessage"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	authToken, err := CreateAuthToken(ctx, params)
	if err != nil {
		fmt.Printf("create auth token failed: %v\n", err)
		return nil, err
	}

	err = CreateMessage(ctx, params, authToken)
	if err != nil {
		fmt.Printf("create message rpc failed: %v\n", err)
		return nil, err
	}

	return env.NewSimpleExec(), nil

}

func CreateAuthToken(ctx context.Context, params TestCaseParams) (string, error) {
	adminToken, err := params.VenusAuth.Param("AdminToken")
	if err != nil {
		return "", err
	}

	endpoint := params.VenusAuth.SvcEndpoint()
	if env.Debug {
		pods, err := params.VenusAuth.Pods(ctx)
		if err != nil {
			return "", err
		}

		svc, err := params.VenusAuth.Svc(ctx)
		if err != nil {
			return "", err
		}
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return "", err
		}
	}

	authAPIClient, err := jwtclient.NewAuthClient(endpoint.ToHttp(), adminToken.(string))
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

	return authAPIClient.GenerateToken(ctx, "admin", "admin", "")
}

func CreateMessage(ctx context.Context, params TestCaseParams, authToken string) error {
	endpoint := params.VenusMessage.SvcEndpoint()
	if env.Debug {
		pods, err := params.VenusAuth.Pods(ctx)
		if err != nil {
			return err
		}

		svc, err := params.VenusAuth.Svc(ctx)
		if err != nil {
			return err
		}
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
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
