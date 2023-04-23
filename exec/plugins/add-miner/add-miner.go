package main

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus/venus-shared/api/wallet"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
	"math/rand"
)

var Info = types.PluginInfo{
	Name:        "add_miner",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "user add miner",
}

type TestCaseParams struct {
	fx.In
	AdminToken                 types.AdminToken
	K8sEnv                     *env.K8sEnvDeployer             `json:"-"`
	VenusWallet                env.IVenusWalletDeployer        `json:"-"`
	VenusSectorManagerDeployer env.IVenusSectorManagerDeployer `json:"-"`
}

func Exec(ctx context.Context, params TestCaseParams) error {

	walletAddr, err := createWallet(ctx, params)
	if err != nil {
		return err
	}

	cmd := []string{
		"./venus-sector-manager",
		"util",
		"miner",
		"create",
		"--sector-size=32GiB",
		"--exid=" + string(rune(rand.Intn(1000))),
	}

	cmd = append(cmd, "--from="+walletAddr.String())
	fmt.Println(cmd)

	minerAddr, err := params.K8sEnv.ExecRemoteCmd(ctx, params.VenusSectorManagerDeployer.Pods()[0].GetName(), cmd)
	if err != nil {
		return err
	}

	getMinerCmd := []string{
		"./venus-sector-manager",
		"util",
		"miner",
		"info",
		string(minerAddr),
	}
	minerInfo, err := params.K8sEnv.ExecRemoteCmd(ctx, params.VenusSectorManagerDeployer.Pods()[0].GetName(), getMinerCmd)
	if err != nil {
		return err
	}
	fmt.Println(minerInfo)
	return nil
}

func createWallet(ctx context.Context, params TestCaseParams) (address.Address, error) {

	walletToken, err := env.ReadWalletToken(ctx, params.K8sEnv, params.VenusWallet.Pods()[0].GetName())
	if err != nil {
		fmt.Println("请提供venus_wallet组件")
		return address.Undef, err
	}

	endpoint := params.VenusWallet.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, params.VenusWallet.Pods()[0].GetName(), int(params.VenusWallet.Svc().Spec.Ports[0].Port))
		if err != nil {
			return address.Undef, err
		}
	}

	walletRpc, closer, err := wallet.DialIFullAPIRPC(ctx, endpoint.ToMultiAddr(), walletToken, nil)
	if err != nil {
		return address.Undef, err
	}
	defer closer()

	password := "123456"
	err = walletRpc.SetPassword(ctx, password)
	if err != nil {
		return address.Undef, err
	}
	walletAddr, err := walletRpc.WalletNew(ctx, vTypes.KTBLS)
	if err != nil {
		return address.Undef, err
	}
	fmt.Println("wallet:", walletAddr)
	return walletAddr, nil
}
