package main

import (
	"context"
	"fmt"

	v2API "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "verity_gateway",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "verity gateway if normal",
}

type TestCaseParams struct {
	fx.In
	Params struct {
		AuthorizerURL string `json:"authorizer_url"`
	} `optional:"true"`
	K8sEnv         *env.K8sEnvDeployer `json:"-"`
	VenusWalletPro env.IDeployer       `json:"-" svcname:"VenusWalletPro"`
	VenusAuth      env.IDeployer       `json:"-" svcname:"VenusAuth"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {

	walletAddrs, err := ImportFbls(ctx, params)
	if err != nil || len(walletAddrs) <= 0 {
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

	adminTokenV, err := params.VenusAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	err = GetWalletInfo(ctx, params, adminTokenV.(string), walletAddrs[0])
	if err != nil {
		fmt.Printf("get wallet info failed: %v\n", err)
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

func GetWalletInfo(ctx context.Context, params TestCaseParams, authToken string, walletAddr string) error {
	endpoint := params.VenusAuth.SvcEndpoint()
	if env.Debug {
		pods, err := params.VenusAuth.Pods(ctx)
		if err != nil {
			return err
		}

		svc, err := params.VenusAuth.Svc(ctx)
		if err != nil {
			return err
		}
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return err
		}
	}

	api, closer, err := v2API.DialIGatewayRPC(ctx, endpoint.ToHttp(), authToken, nil)
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
