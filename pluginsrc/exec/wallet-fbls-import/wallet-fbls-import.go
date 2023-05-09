package main

import (
	"context"
	"fmt"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

// Info
var Info = types.PluginInfo{
	Name:        "wallet-fbls-import",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "wallet-fbls-import",
}

// TestCaseParams
type TestCaseParams struct {
	fx.In
	K8sEnv         *env.K8sEnvDeployer `json:"-"`
	VenusWalletPro env.IDeployer       `json:"-" svcname:"VenusWalletPro"`
}

// Exec
func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {

	walletAddrs, err := ImportFbls(ctx, params)
	if err != nil {
		fmt.Printf("create miner failed: %v\n", err)
		return nil, err
	}
	for id, addr := range walletAddrs {
		fmt.Println("wallet %v is: %v", id, addr)
	}

	return env.NewSimpleExec(), nil
}

// ImportFbls
func ImportFbls(ctx context.Context, params TestCaseParams) ([]string, error) {
	venusWalletProPods, err := params.VenusWalletPro.Pods(ctx)
	if err != nil {
		return nil, err
	}
	cmd := []string{
		"./venus-wallet-pro",
		"wallet",
		"fbls_import",
		"--file",
		"/root/fbls.key",
	}

	var addrs []string

	walletAaddrs, err := params.K8sEnv.ExecRemoteCmd(ctx, venusWalletProPods[0].GetName(), cmd...)
	if err != nil {
		return nil, fmt.Errorf("exec remote cmd failed: %w\n", err)
	}

	for _, b := range walletAaddrs {
		addrs = append(addrs, string(b))
	}

	return addrs, nil
}
