package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/go-address"
	v2API "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	"github.com/filecoin-project/venus/venus-shared/api/wallet"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
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
	K8sEnv       *env.K8sEnvDeployer `json:"-"`
	VenusGateway env.IDeployer       `json:"-" svcname:"VenusGateway"`
	VenusWallet  env.IDeployer       `json:"-" svcname:"VenusWallet"`
	VenusAuth    env.IDeployer       `json:"-" svcname:"VenusAuth"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {

	_, err := CreateWallet(ctx, params)
	if err != nil {
		fmt.Printf("create wallet failed: %v\n", err)
		return nil, err
	}

	adminTokenV, err := params.VenusAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	err = GetWalletInfo(ctx, params, adminTokenV.(string))
	if err != nil {
		fmt.Printf("get wallet info failed: %v\n", err)
		return nil, err
	}

	return env.NewSimpleExec(), nil

}

func CreateWallet(ctx context.Context, params TestCaseParams) (address.Address, error) {
	pods, err := params.VenusWallet.Pods(ctx)
	if err != nil {
		return address.Undef, err
	}

	svc, err := params.VenusWallet.Svc(ctx)
	if err != nil {
		return address.Undef, err
	}

	walletToken, err := env.ReadWalletToken(ctx, params.K8sEnv, pods[0].GetName())
	if err != nil {
		return address.Undef, fmt.Errorf("read wallet token failed: %w\n", err)
	}

	endpoint := params.VenusWallet.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return address.Undef, fmt.Errorf("port forward failed: %w\n", err)
		}
	}

	walletRpc, closer, err := wallet.DialIFullAPIRPC(ctx, endpoint.ToMultiAddr(), walletToken, nil)
	if err != nil {
		return address.Undef, fmt.Errorf("dial iFullAPI rpc failed: %w\n", err)
	}
	defer closer()

	password := "123456"
	err = walletRpc.SetPassword(ctx, password)
	if err != nil {
		return address.Undef, fmt.Errorf("set password failed: %w\n", err)
	}

	walletAddr, err := walletRpc.WalletNew(ctx, vTypes.KTBLS)
	if err != nil {
		return address.Undef, fmt.Errorf("create wallet failed: %w\n", err)
	}
	fmt.Printf("wallet: %v\n", walletAddr)

	return walletAddr, nil
}

func GetWalletInfo(ctx context.Context, params TestCaseParams, authToken string) error {
	endpoint := params.VenusWallet.SvcEndpoint()
	if env.Debug {
		pods, err := params.VenusWallet.Pods(ctx)
		if err != nil {
			return err
		}

		svc, err := params.VenusWallet.Svc(ctx)
		if err != nil {
			return err
		}
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return err
		}
	}

	api, closer, err := v2API.DialIGatewayRPC(ctx, endpoint.ToHttp(), authToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	wallets, err := api.ListWalletInfo(ctx)
	if err != nil {
		return err
	}
	minersBytes, err := json.MarshalIndent(wallets, " ", "\t")
	if err != nil {
		return err
	}
	fmt.Println(string(minersBytes))
	return nil
}
