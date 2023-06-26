package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	venusutils "github.com/hunjixin/brightbird/env/venus_utils"

	"github.com/filecoin-project/venus/venus-shared/api/wallet"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	types2 "github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types2.PluginInfo{
	Name:        "import_key",
	Version:     version.Version(),
	PluginType:  types2.TestExec,
	Description: "import private key to venus wallet",
}

type TestCaseParams struct {
	fx.In
	Params struct {
		PrivKey string `json:"privKey"`
	} `optional:"true"`
	K8sEnv      *env.K8sEnvDeployer `json:"-"`
	VenusWallet env.IDeployer       `json:"-" svcname:"VenusWallet"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	venusWallethPods, err := params.VenusWallet.Pods(ctx)
	if err != nil {
		return nil, err
	}

	svc, err := params.VenusWallet.Svc(ctx)
	if err != nil {
		return nil, err
	}
	walletToken, err := venusutils.ReadWalletToken(ctx, params.K8sEnv, venusWallethPods[0].GetName())
	if err != nil {
		return nil, err
	}

	endpoint, err := params.VenusWallet.SvcEndpoint()
	if err != nil {
		return nil, err
	}
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, venusWallethPods[0].GetName(), int(svc.Spec.Ports[0].Port))
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

	keyBytes, err := hex.DecodeString(params.Params.PrivKey)
	if err != nil {
		return nil, err
	}
	fmt.Println("aaaaaa", string(keyBytes))
	var ki vTypes.KeyInfo
	err = json.Unmarshal(keyBytes, &ki)
	if err != nil {
		return nil, err
	}
	addr, err := walletRPC.WalletImport(ctx, &ki)
	if err != nil {
		return nil, err
	}
	fmt.Println("import key: ", addr)
	return env.NewSimpleExec(), nil
}
