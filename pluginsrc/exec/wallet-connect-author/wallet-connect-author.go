package main

import (
	"context"
	"fmt"
	"go.uber.org/fx"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

// Info
var Info = types.PluginInfo{
	Name:        "wallet-connect-author",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "wallet connect author",
}

// TestCaseParams
type TestCaseParams struct {
	fx.In
	Params struct {
		AuthorizerURL string `json:"authorizer_url"`
	} `optional:"true"`
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

	err = ConnectAuthor(ctx, params)
	if err != nil {
		return nil, err
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

// ImportFbls
func ConnectAuthor(ctx context.Context, params TestCaseParams) error {
	venusWalletProPods, err := params.VenusWalletPro.Pods(ctx)
	if err != nil {
		return err
	}
	cmd := []string{
		"/venus-wallet-pro",
		"wallet",
		"connect_author",
		"--authorizer",
		params.Params.AuthorizerURL,
	}

	_, err = params.K8sEnv.ExecRemoteCmdWithName(ctx, venusWalletProPods[0].GetName(), cmd...)
	if err != nil {
		return fmt.Errorf("exec remote cmd failed: %w\n", err)
	}

	return nil
}
