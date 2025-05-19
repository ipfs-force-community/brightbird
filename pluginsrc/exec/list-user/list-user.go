package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ipfs-force-community/sophon-auth/core"
	"github.com/ipfs-force-community/sophon-auth/jwtclient"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"

	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "list_user",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "list user",
}

type TestCaseParams struct {
	Skip     string                            `json:"skip" jsonschema:"skip" title:"PageSkip" require:"true"`
	Limit    string                            `json:"limit" jsonschema:"limit" title:"PageLimit" require:"true"`
	UserName string                            `json:"userName" jsonschema:"userName" title:"userName" require:"true" description:"user name, if empty not check"`
	Auth     sophonauth.SophonAuthDeployReturn `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	authAPIClient, err := jwtclient.NewAuthClient(params.Auth.SvcEndpoint.ToHTTP(), params.Auth.AdminToken)
	if err != nil {
		return err
	}

	skip, err := strconv.ParseInt(params.Skip, 10, 64)
	if err != nil {
		return err
	}

	limit, err := strconv.ParseInt(params.Limit, 10, 64)
	if err != nil {
		return err
	}
	userList, err := authAPIClient.ListUsers(ctx, skip, limit, core.UserStateUndefined)
	if err != nil {
		return err
	}

	hassUser := false
	fmt.Println("-------list-users-------")
	for _, user := range userList {
		fmt.Printf("%+v\n", *user)
		if params.UserName == user.Name {
			hassUser = true
		}
	}
	if len(params.UserName) > 0 {
		if !hassUser {
			return fmt.Errorf("not find user : %v", params.UserName)
		} else {
			fmt.Println("got user: ", params.UserName)
		}
	}

	return nil
}
