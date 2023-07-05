package main

import (
	"context"
	"fmt"

	v2API "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	sophongateway "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-gateway"
	venuswallet "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-wallet"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "check_user_in_gateway",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "check whether usr wallet has register in gateway",
}

type TestCaseParams struct {
	SophonAuth  sophonauth.SophonAuthDeployReturn `json:"SophonAuth"`
	Gateway     sophongateway.SophonGatewayReturn `json:"Gateway"`
	VenusWallet venuswallet.VenusWalletReturn     `json:"VenusWallet"`
	UserName    string                            `json:"userName"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	api, closer, err := v2API.DialIGatewayRPC(ctx, params.Gateway.SvcEndpoint.ToMultiAddr(), params.SophonAuth.AdminToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	wallets, err := api.ListWalletInfo(ctx)
	if err != nil {
		return err
	}
	for _, wallet := range wallets {
		if wallet.Account == params.UserName {
			return nil
		}
	}
	return fmt.Errorf("user wallet %s not found", params.UserName)
}
