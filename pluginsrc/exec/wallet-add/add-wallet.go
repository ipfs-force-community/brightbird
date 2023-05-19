package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/venus/venus-shared/api/wallet"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "add_wallet",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "add wallet",
}

type TestCaseParams struct {
	fx.In
	K8sEnv      *env.K8sEnvDeployer `json:"-"`
	VenusAuth   env.IDeployer       `json:"-" svcname:"VenusAuth"`
	VenusWallet env.IDeployer       `json:"-" svcname:"VenusWallet"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {

	adminTokenV, err := params.VenusAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	walletAddr, err := CreateWallet(ctx, params, adminTokenV.MustString())
	if err != nil {
		fmt.Printf("create wallet failed: %v\n", err)
		return nil, err
	}

	return env.NewSimpleExec().Add("Wallet", env.ParamsFromVal(walletAddr)), nil
}

func CreateWallet(ctx context.Context, params TestCaseParams, token string) (string, error) {
	pods, err := params.VenusWallet.Pods(ctx)
	if err != nil {
		return "", err
	}

	svc, err := params.VenusWallet.Svc(ctx)
	if err != nil {
		return "", err
	}

	endpoint, err := params.VenusWallet.SvcEndpoint()
	if err != nil {
		return "", err
	}

	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return "", fmt.Errorf("port forward failed: %w\n", err)
		}
	}

	walletRpc, closer, err := wallet.DialIFullAPIRPC(ctx, endpoint.ToMultiAddr(), token, nil)
	if err != nil {
		return "", fmt.Errorf("dial iFullAPI rpc failed: %w\n", err)
	}
	defer closer()

	password := "123456"
	err = walletRpc.SetPassword(ctx, password)
	if err != nil {
		return "", fmt.Errorf("set password failed: %w\n", err)
	}

	walletAddr, err := walletRpc.WalletNew(ctx, vTypes.KTBLS)
	if err != nil {
		return "", fmt.Errorf("create wallet failed: %w\n", err)
	}
	fmt.Printf("wallet: %v\n", walletAddr)

	return walletAddr.String(), nil
}
