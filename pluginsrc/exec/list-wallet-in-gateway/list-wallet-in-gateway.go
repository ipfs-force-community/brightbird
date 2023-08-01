package main

import (
	"context"

	v2API "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	"github.com/filecoin-project/venus/venus-shared/types/gateway"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	sophongateway "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-gateway"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "get-wallet-in-gateway",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "list wallet and address in gateway",
}

type TestCaseParams struct {
	Auth     sophonauth.SophonAuthDeployReturn `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Gateway  sophongateway.SophonGatewayReturn `json:"SophonGateway"  jsonschema:"SophonGateway"  title:"SophonGateway" require:"true" description:"gateway deploy return"`
	UserName string                            `json:"userName" jsonschema:"userName" title:"UserName" require:"true" description:"user name"`
}

// todo support array
type ListWalletInGatewayReturn = gateway.WalletDetail

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*ListWalletInGatewayReturn, error) {
	api, closer, err := v2API.DialIGatewayRPC(ctx, params.Gateway.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	return api.ListWalletInfoByWallet(ctx, params.UserName)
}
