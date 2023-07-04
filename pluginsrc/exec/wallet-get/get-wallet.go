package main

import (
	"context"
	"fmt"

	v2API "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "get_wallet",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "get wallet",
}

type TestCaseParams struct {
	fx.In
	K8sEnv     *env.K8sEnvDeployer `json:"-"`
	SophonAuth env.IDeployer       `json:"-" svcname:"SophonAuth"`
	AddWallet  env.IExec           `json:"-" svcname:"AddWallet"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {

	adminTokenV, err := params.SophonAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	walletAddr, err := params.AddWallet.Param("Wallet")
	if err != nil {
		return nil, err
	}

	err = GetWalletInfo(ctx, params, adminTokenV.MustString(), walletAddr.MustString())
	if err != nil {
		fmt.Printf("get wallet info failed: %v\n", err)
		return nil, err
	}

	return env.NewSimpleExec().Add("Wallet", walletAddr), nil
}

func GetWalletInfo(ctx context.Context, params TestCaseParams, authToken string, walletAddr string) error {
	endpoint, err := params.SophonAuth.SvcEndpoint()
	if err != nil {
		return err
	}

	api, closer, err := v2API.DialIGatewayRPC(ctx, endpoint.ToHTTP(), authToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	walletDetail, err := api.ListWalletInfoByWallet(ctx, walletAddr)
	if err != nil {
		return err
	}

	fmt.Println(walletDetail)
	return nil
}
