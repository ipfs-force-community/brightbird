package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"github.com/ipfs-force-community/sophon-auth/jwtclient"
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

type CreateTokenReturn struct {
	Token string `json:"token" description:"generated token"`
}

type TestCaseParams struct {
	SophonAuth sophonauth.SophonAuthDeployReturn `json:"SophonAuth" description:"[Deploy]sophon-auth"`

	UserName string `json:"UserName" description:"token user name"`
	Perm     string `json:"perm" description:"[OPTIONS] custom string in JWT payload"`
	Extra    string `json:"extra" description:"[OPTIONS] permission for API auth (read, write, sign, admin)"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*CreateTokenReturn, error) {
	authAPIClient, err := jwtclient.NewAuthClient(params.SophonAuth.SvcEndpoint.ToHTTP(), params.SophonAuth.AdminToken)
	if err != nil {
		return nil, err
	}

	token, err := authAPIClient.GenerateToken(ctx, params.UserName, params.Perm, params.Extra)
	if err != nil {
		return nil, err
	}

	return &CreateTokenReturn{
		Token: token,
	}, nil
}
