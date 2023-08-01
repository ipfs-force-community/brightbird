package main

import (
	"context"
	"encoding/hex"
	"encoding/json"

	venusutils "github.com/ipfs-force-community/brightbird/env/venus_utils"
	venuswallet "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus-wallet"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus/venus-shared/api/wallet"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	types2 "github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
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

type ImportKeyReturn struct {
	Address address.Address `json:"address" jsonschema:"address" title:"Address" description:"import address" require:"true"`
}

type TestCaseParams struct {
	PrivKey     string                        `json:"privKey" jsonschema:"privKey" title:"Private Key" require:"true" description:"private key for venus/lotus keyinfo "`
	VenusWallet venuswallet.VenusWalletReturn `json:"VenusWallet" jsonschema:"VenusWallet" title:"Venus Wallet" description:"wallet return" require:"true"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*ImportKeyReturn, error) {
	venusWallethPods, err := venuswallet.GetPods(ctx, k8sEnv, params.VenusWallet.InstanceName)
	if err != nil {
		return nil, err
	}
	walletToken, err := venusutils.ReadWalletToken(ctx, k8sEnv, venusWallethPods[0].GetName())
	if err != nil {
		return nil, err
	}

	walletRPC, closer, err := wallet.DialIFullAPIRPC(ctx, params.VenusWallet.SvcEndpoint.ToMultiAddr(), walletToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	keyBytes, err := hex.DecodeString(params.PrivKey)
	if err != nil {
		return nil, err
	}
	var ki vTypes.KeyInfo
	err = json.Unmarshal(keyBytes, &ki)
	if err != nil {
		return nil, err
	}
	addr, err := walletRPC.WalletImport(ctx, &ki)
	if err != nil {
		return nil, err
	}
	return &ImportKeyReturn{
		Address: addr,
	}, nil
}
