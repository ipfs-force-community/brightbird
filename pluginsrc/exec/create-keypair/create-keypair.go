package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	venusutils "github.com/hunjixin/brightbird/env/venus_utils"
	venuswallet "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-wallet"

	"github.com/filecoin-project/venus/venus-shared/api/wallet"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "create_keypair",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "create key pair",
}

type TestCaseParams struct {
	VenusWallet venuswallet.VenusWalletReturn `json:"VenusWallet"`
	KeyType     string                        `json:"keyType" description:"private key type bls/secp256k1/delegated"`
}

type CreateKeyPair struct {
	Address    string
	PrivateKey string
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*CreateKeyPair, error) {
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

	walletAddr, err := walletRPC.WalletNew(ctx, vTypes.KeyType(params.KeyType))
	if err != nil {
		return nil, fmt.Errorf("create wallet failed: %w", err)
	}

	keyInfo, err := walletRPC.WalletExport(ctx, walletAddr)
	if err != nil {
		return nil, fmt.Errorf("create wallet failed: %w", err)
	}

	kiBytes, err := json.Marshal(keyInfo)
	if err != nil {
		return nil, err
	}

	return &CreateKeyPair{
		Address:    walletAddr.String(),
		PrivateKey: hex.EncodeToString(kiBytes),
	}, nil
}
