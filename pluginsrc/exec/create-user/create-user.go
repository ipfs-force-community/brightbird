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

	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
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
	SophanAuthDeploy sophonauth.SophonAuthDeployReturn `json:"SophanAuth"`
	UserName         string                            `json:"userName"`
	Comment          string                            `json:"comment"`
}

type CreateUserReturn struct {
	UserName string `json:"userName"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*CreateUserReturn, error) {
	authAPIClient, err := jwtclient.NewAuthClient(params.SophanAuthDeploy.SvcEndpoint.ToHTTP(), params.SophanAuthDeploy.AdminToken)
	if err != nil {
		return nil, err
	}

	if len(params.UserName) == 0 {
		return nil, fmt.Errorf("username cant be empty")
	}

	user, err := authAPIClient.CreateUser(ctx, &auth.CreateUserRequest{
		Name:    params.UserName,
		Comment: &params.Comment,
		State:   core.UserStateEnabled,
	})
	if err != nil {
		return nil, err
	}

	fmt.Println(user.Name)
	return &CreateUserReturn{
		UserName: user.Name,
	}, nil
}
