package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/venus/venus-shared/api/wallet"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
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
	Name:        "wallet_new",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "wallet_new",
}

type TestCaseParams struct {
	fx.In
	K8sEnv      *env.K8sEnvDeployer `json:"-"`
	VenusWallet env.IDeployer       `json:"-" svcname:"VenusWallet"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {

	walletPods, err := params.VenusWallet.Pods(ctx)
	if err != nil {
		return nil, err
	}

	walletToken, err := env.ReadWalletToken(ctx, params.K8sEnv, walletPods[0].GetName())
	if err != nil {
		return nil, err
	}

	endpoint, err := params.VenusWallet.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	if env.Debug {
		svc, err := params.VenusWallet.Svc(ctx)
		if err != nil {
			return nil, err
		}
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, walletPods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return nil, err
		}
	}

	walletRPC, closer, err := wallet.DialIFullAPIRPC(ctx, endpoint.ToMultiAddr(), walletToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	walletAddr, err := walletRPC.WalletNew(ctx, vTypes.KTBLS)
	if err != nil {
		return nil, fmt.Errorf("create wallet failed: %w", err)
	}
	fmt.Printf("wallet: %v\n", walletAddr)

	return env.NewSimpleExec().Add("Wallet", env.ParamsFromVal(walletAddr)), nil
}
