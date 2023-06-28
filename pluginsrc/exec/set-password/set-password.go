package main

import (
	"context"
	"fmt"

	venusutils "github.com/hunjixin/brightbird/env/venus_utils"

	"github.com/filecoin-project/venus/venus-shared/api/wallet"
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
	Name:        "set_password",
	Version:     version.Version(),
	PluginType:  types.TestExec,
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

	walletToken, err := venusutils.ReadWalletToken(ctx, params.K8sEnv, walletPods[0].GetName())
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

	version, err := walletRPC.Version(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println("wallet:", version)

	err = walletRPC.SetPassword(ctx, params.Params.Password)
	if err != nil {
		return nil, err
	}
	return env.NewSimpleExec(), nil
}
