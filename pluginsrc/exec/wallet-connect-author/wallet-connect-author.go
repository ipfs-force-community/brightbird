package main

import (
	"context"
	"fmt"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	venuswalletpro "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-wallet-pro"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("wallet-connect-author")

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
	AuthorizerURL  string                                    `json:"authorizerUrl" jsonschema:"authorizerUrl" title:"AuthorizerUrl" require:"true" description:"wallet pro auth url"`
	VenusWalletPro venuswalletpro.VenusWalletProDeployReturn `json:"VenusWalletPro"  jsonschema:"VenusWalletPro" title:"Venus Wallet Auth" require:"true" description:"venus wallet return"`
}

// Exec
func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	walletAddrs, err := ImportFbls(ctx, k8sEnv, params)
	if err != nil {
		fmt.Printf("create miner failed: %v\n", err)
		return err
	}
	for id, addr := range walletAddrs {
		log.Infof("wallet %v is: %v", id, addr)
	}

	err = ConnectAuthor(ctx, k8sEnv, params)
	if err != nil {
		return err
	}

	return nil
}

// ImportFbls
func ImportFbls(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) ([]string, error) {
	venusWalletProPods, err := venuswalletpro.GetPods(ctx, k8sEnv, params.VenusWalletPro.InstanceName)
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

	walletAaddrs, err := k8sEnv.ExecRemoteCmd(ctx, venusWalletProPods[0].GetName(), cmd...)
	if err != nil {
		return nil, fmt.Errorf("exec remote cmd failed: %w", err)
	}

	for _, b := range walletAaddrs {
		addrs = append(addrs, string(b))
	}

	return addrs, nil
}

// ImportFbls
func ConnectAuthor(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	venusWalletProPods, err := venuswalletpro.GetPods(ctx, k8sEnv, params.VenusWalletPro.InstanceName)
	if err != nil {
		return err
	}
	cmd := []string{
		"/venus-wallet-pro",
		"wallet",
		"connect_author",
		"--authorizer",
		params.AuthorizerURL,
	}

	_, err = k8sEnv.ExecRemoteCmdWithName(ctx, venusWalletProPods[0].GetName(), cmd...)
	if err != nil {
		return fmt.Errorf("exec remote cmd failed: %w", err)
	}

	return nil
}
