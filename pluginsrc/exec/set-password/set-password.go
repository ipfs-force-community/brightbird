package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/venus/venus-shared/api/wallet"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "set_password",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "import private key to venus wallet",
}

type TestCaseParams struct {
	fx.In
	Params struct {
		Password string `json:"password" optional:"true"`
	} `optional:"true"`
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

	endpoint := params.VenusWallet.SvcEndpoint()
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

	walletRpc, closer, err := wallet.DialIFullAPIRPC(ctx, endpoint.ToMultiAddr(), walletToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	version, err := walletRpc.Version(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println("wallet:", version)

	err = walletRpc.SetPassword(ctx, params.Params.Password)
	if err != nil {
		return nil, err
	}
	addr, err := walletRpc.WalletNew(ctx, vTypes.KTBLS)
	if err != nil {
		return nil, err
	}
	fmt.Println("wallet:", addr)
	return env.NewSimpleExec(), nil
}
