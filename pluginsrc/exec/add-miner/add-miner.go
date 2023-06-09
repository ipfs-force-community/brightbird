package main

import (
	"context"
	"fmt"
	"math/rand"

	"go.uber.org/fx"

	"github.com/filecoin-project/go-address"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("add-miner")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "add_miner",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "user add miner",
}

type TestCaseParams struct {
	fx.In

	K8sEnv                     *env.K8sEnvDeployer `json:"-"`
	VenusWallet                env.IDeployer       `json:"-" svcname:"VenusWallet"`
	VenusSectorManagerDeployer env.IDeployer       `json:"-" svcname:"VenusSectorManager"`
	CreateWallet               env.IExec           `json:"-" svcname:"CreateWallet"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	walletAddr, err := params.CreateWallet.Param("Wallet")
	if err != nil {
		return nil, err
	}

	addr, err := env.UnmarshalJSON[address.Address](walletAddr.Raw())
	if err != nil {
		panic(err)
	}

	minerAddr, err := CreateMiner(ctx, params, addr)
	if err != nil {
		fmt.Printf("create miner failed: %v\n", err)
		return nil, err
	}

	minerInfo, err := GetMinerInfo(ctx, params, minerAddr)
	if err != nil {
		fmt.Printf("get miner info failed: %v\n", err)
		return nil, err
	}

	log.Infof("miner info: %v", minerInfo)
	return env.NewSimpleExec().Add("Miner", env.ParamsFromVal(minerAddr)), nil
}

func CreateMiner(ctx context.Context, params TestCaseParams, walletAddr address.Address) (string, error) {
	venusWalletPods, err := params.VenusWallet.Pods(ctx)
	if err != nil {
		return "", err
	}
	cmd := []string{
		"./venus-sector-manager",
		"util",
		"miner",
		"create",
		"--sector-size=8MiB",
		"--exid=" + string(rune(rand.Intn(100000))),
	}
	cmd = append(cmd, "--from="+walletAddr.String())

	minerAddr, err := params.K8sEnv.ExecRemoteCmd(ctx, venusWalletPods[0].GetName(), cmd...)
	if err != nil {
		return "", fmt.Errorf("exec remote cmd failed: %w", err)
	}

	return string(minerAddr), nil
}

func GetMinerInfo(ctx context.Context, params TestCaseParams, minerAddr string) (string, error) {
	venusWalletPods, err := params.VenusWallet.Pods(ctx)
	if err != nil {
		return "", err
	}
	getMinerCmd := []string{
		"./venus-sector-manager",
		"util",
		"miner",
		"info",
		minerAddr,
	}
	minerInfo, err := params.K8sEnv.ExecRemoteCmd(ctx, venusWalletPods[0].GetName(), getMinerCmd...)
	if err != nil {
		return "", fmt.Errorf("exec remote cmd failed: %w", err)
	}

	return string(minerInfo), nil
}
