package venusutils

import (
	"context"
	"encoding/hex"

	"github.com/ipfs-force-community/brightbird/env"

	"github.com/BurntSushi/toml"
	walletCfg "github.com/filecoin-project/venus-wallet/config"
)

func ReadWalletToken(ctx context.Context, k8sEnv *env.K8sEnvDeployer, walletPod string) (string, error) {
	cfgBytes, err := k8sEnv.ReadSmallFilelInPod(ctx, walletPod, "/root/.venus_wallet/config.toml")
	if err != nil {
		return "", err
	}
	walletCfg := new(walletCfg.Config)
	err = toml.Unmarshal(cfgBytes, walletCfg)
	if err != nil {
		return "", err
	}

	tokenBytes, err := hex.DecodeString(walletCfg.JWT.Token)
	if err != nil {
		return "", err
	}

	return string(tokenBytes), nil
}
