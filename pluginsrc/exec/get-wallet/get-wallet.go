package main

import (
	"context"
	"fmt"

	v2API "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "verity_gateway",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "verity gateway if normal",
}

type TestCaseParams struct {
	fx.In
	Params struct {
		AuthorizerURL string `json:"authorizer_url"`
	} `optional:"true"`
	K8sEnv         *env.K8sEnvDeployer `json:"-"`
	VenusWalletPro env.IDeployer       `json:"-" svcname:"VenusWalletPro"`
	VenusAuth      env.IDeployer       `json:"-" svcname:"VenusAuth"`
	WalletName     env.IExec           `json:"-" svcname:"WalletName"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {

	adminTokenV, err := params.VenusAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	walletAddr, err := params.WalletName.Param("WalletName")

	_, err = GetWalletInfo(ctx, params, adminTokenV.(string), walletAddr.(string))
	if err != nil {
		fmt.Printf("get wallet info failed: %v\n", err)
		return nil, err
	}

	return env.NewSimpleExec().Add("Wallet", walletAddr), nil
}

func GetWalletInfo(ctx context.Context, params TestCaseParams, authToken string, walletAddr string) (env.IExec, error) {
	endpoint := params.VenusAuth.SvcEndpoint()
	if env.Debug {
		pods, err := params.VenusAuth.Pods(ctx)
		if err != nil {
			return nil, err
		}

		svc, err := params.VenusAuth.Svc(ctx)
		if err != nil {
			return nil, err
		}
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return nil, err
		}
	}

	api, closer, err := v2API.DialIGatewayRPC(ctx, endpoint.ToHttp(), authToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	walletDetail, err := api.ListWalletInfoByWallet(ctx, walletAddr)
	if err != nil {
		return nil, err
	}

	fmt.Println(walletDetail)
	return env.NewSimpleExec().Add("WalletInfo", walletAddr), nil
}
