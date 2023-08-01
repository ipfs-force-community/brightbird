package main

import (
	"context"
	"fmt"

	venusutils "github.com/ipfs-force-community/brightbird/env/venus_utils"
	venuswallet "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus-wallet"

	"github.com/filecoin-project/venus/venus-shared/api/wallet"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
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
	Password    string                        `json:"password" jsonschema:"password" title:"Password" require:"true" description:"set password to wallet"`
	VenusWallet venuswallet.VenusWalletReturn `json:"VenusWallet" jsonschema:"VenusWallet" title:"Venus Wallet" description:"wallet return" require:"true"`
}
type SetPasswordReturn struct {
	Password string `json:"password" jsonschema:"password" title:"Password" require:"true" description:"password for venus wallet"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*SetPasswordReturn, error) {
	walletPods, err := venuswallet.GetPods(ctx, k8sEnv, params.VenusWallet.InstanceName)
	if err != nil {
		return nil, err
	}

	walletToken, err := venusutils.ReadWalletToken(ctx, k8sEnv, walletPods[0].GetName())
	if err != nil {
		return nil, err
	}

	walletRPC, closer, err := wallet.DialIFullAPIRPC(ctx, params.VenusWallet.SvcEndpoint.ToMultiAddr(), walletToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	version, err := walletRPC.Version(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println("wallet:", version)

	err = walletRPC.SetPassword(ctx, params.Password)
	if err != nil {
		return nil, err
	}
	return &SetPasswordReturn{
		Password: params.Password,
	}, nil
}
