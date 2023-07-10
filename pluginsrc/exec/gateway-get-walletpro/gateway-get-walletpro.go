package main

import (
	"context"
	"fmt"

	v2API "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	venuswalletpro "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-wallet-pro"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("walletpro_gateway")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "walletpro_gateway",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "walletpro_gateway",
}

type TestCaseParams struct {
	AuthorizerURL  string                                    `json:"authorizerUrl" jsonschema:"authorizerUrl" title:"AuthorizerUrl" require:"true" description:"wallet pro auth url"`
	Auth           sophonauth.SophonAuthDeployReturn         `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	VenusWalletPro venuswalletpro.VenusWalletProDeployReturn `json:"VenusWalletPro"  jsonschema:"VenusWalletPro" title:"Venus Wallet Auth" require:"true" description:"venus wallet return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	walletAddrs, err := ImportFbls(ctx, k8sEnv, params)
	if err != nil || len(walletAddrs) <= 0 {
		return fmt.Errorf("create miner err %w", err)
	}

	err = ConnectAuthor(ctx, k8sEnv, params)
	if err != nil {
		return err
	}

	err = GetWalletInfo(ctx, k8sEnv, params, params.Auth.AdminToken, walletAddrs[0])
	if err != nil {
		return fmt.Errorf("get wallet finfo failed %w", err)
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

func GetWalletInfo(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams, authToken string, walletAddr string) error {
	api, closer, err := v2API.DialIGatewayRPC(ctx, params.Auth.SvcEndpoint.ToHTTP(), authToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	walletDetail, err := api.ListWalletInfoByWallet(ctx, walletAddr)
	if err != nil {
		return err
	}

	fmt.Println(walletDetail)
	return nil
}
