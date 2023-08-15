package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	v2API "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	sophongateway "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-gateway"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
	"github.com/ipfs-force-community/sophon-auth/jwtclient"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "user-miner-add",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "user miner add",
}

type TestCaseParams struct {
	Auth     sophonauth.SophonAuthDeployReturn `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Gateway  sophongateway.SophonGatewayReturn `json:"SophonGateway"  jsonschema:"SophonGateway"  title:"SophonGateway" require:"true" description:"gateway deploy return"`
	UserName string                            `json:"userName" jsonschema:"userName" title:"UserName" require:"true" description:"user name"`
	Miner    address.Address                   `json:"miner" jsonschema:"miner" title:"Miner Address" require:"true" description:"miner address"`
}

type CreateUserReturn struct {
	UserName string `json:"userName" jsonschema:"userName" title:"UserName" require:"true" description:"user name in sophon auth"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	authAPIClient, err := jwtclient.NewAuthClient(params.Auth.SvcEndpoint.ToHTTP(), params.Auth.AdminToken)
	if err != nil {
		return err
	}

	if len(params.UserName) == 0 {
		return fmt.Errorf("username cant be empty")
	}

	openMining := true
	var isCreate bool
	if isCreate, err = authAPIClient.UpsertMiner(ctx, params.UserName, params.Miner.String(), openMining); err != nil {
		return err
	}
	var opStr string
	if isCreate {
		opStr = "create"
	} else {
		opStr = "update"
	}
	fmt.Printf("%s user:%s miner:%s success.\n", opStr, params.UserName, params.Miner)

	err = getMinerFromGateway(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func getMinerFromGateway(ctx context.Context, params TestCaseParams) error {
	api, closer, err := v2API.DialIGatewayRPC(ctx, params.Gateway.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	minerList, err := api.ListConnectedMiners(ctx)
	if err != nil {
		return err
	}
	for _, miner := range minerList {
		if miner == params.Miner {
			return nil
		}
	}
	return fmt.Errorf("miner %s not found", params.Miner)
}
