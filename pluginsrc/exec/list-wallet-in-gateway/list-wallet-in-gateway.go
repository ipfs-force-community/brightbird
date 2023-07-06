package main

import (
	"context"

	v2API "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	"github.com/filecoin-project/venus/venus-shared/types/gateway"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	sophongateway "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-gateway"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
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
	SophonAuth sophonauth.SophonAuthDeployReturn `json:"SophonAuth"`
	Gateway    sophongateway.SophonGatewayReturn `json:"Gateway"`
	UserName   string                            `json:"userName"`
}

// todo support arrary
type ListWalletInGatewayReturn = gateway.WalletDetail

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	api, closer, err := v2API.DialIGatewayRPC(ctx, params.Gateway.SvcEndpoint.ToMultiAddr(), params.SophonAuth.AdminToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	_, err = api.ListWalletInfoByWallet(ctx, params.UserName)
	return err
}
