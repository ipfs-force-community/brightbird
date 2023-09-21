package main

import (
	"context"
	"fmt"

	venusutils "github.com/ipfs-force-community/brightbird/env/venus_utils"
	venuswallet "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus-wallet"
	"github.com/ipfs-force-community/sophon-auth/core"

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
	Name:        "create-wallet-auth-token",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "创建wallet token",
}

type TestCaseParams struct {
	VenusWallet venuswallet.VenusWalletReturn `json:"VenusWallet" jsonschema:"VenusWallet" title:"Venus Wallet" description:"wallet return" require:"true"`
	Perm        string                        `json:"perm" jsonschema:"perm" title:"Perm" require:"true" default:"sign" description:"one of: read, write, sign, admin"`
}
type CreateWalletAuthTokenReturn struct {
	Token string `json:"token" jsonschema:"token" title:"Auth Token" require:"true" description:"wallet token generated"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*CreateWalletAuthTokenReturn, error) {
	walletPods, err := venuswallet.GetPods(ctx, k8sEnv, params.VenusWallet.InstanceName)
	if err != nil {
		return nil, err
	}

	walletToken, err := venusutils.ReadWalletToken(ctx, k8sEnv, walletPods[0].GetName())
	if err != nil {
		return nil, err
	}

	api, closer, err := wallet.DialIFullAPIRPC(ctx, params.VenusWallet.SvcEndpoint.ToMultiAddr(), walletToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	allPermissions := core.AdaptOldStrategy(core.PermAdmin)
	perm := params.Perm
	idx := 0
	for i, p := range allPermissions {
		if perm == p {
			idx = i + 1
		}
	}

	if idx == 0 {
		return nil, fmt.Errorf("perm has to be one of: %s", allPermissions)
	}

	token, err := api.AuthNew(ctx, allPermissions[:idx])
	if err != nil {
		return nil, err
	}
	return &CreateWalletAuthTokenReturn{
		Token: string(token),
	}, nil
}
