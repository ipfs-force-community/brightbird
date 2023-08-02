package main

import (
	"context"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
	"github.com/ipfs-force-community/sophon-auth/jwtclient"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "get_user",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "get user name",
}

type TestCaseParams struct {
	UserName string                            `json:"userName" jsonschema:"userName" title:"UserName" require:"true" description:"user name"`
	Auth     sophonauth.SophonAuthDeployReturn `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	authAPIClient, err := jwtclient.NewAuthClient(params.Auth.SvcEndpoint.ToHTTP(), params.Auth.AdminToken)
	if err != nil {
		return err
	}

	_, err = authAPIClient.GetUser(ctx, params.UserName)
	if err != nil {
		return err
	}

	return nil
}
