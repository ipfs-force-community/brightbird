package main

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	v2API "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
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
		log.Infof("create miner failed: %v", err)
		return nil, err
	}
	for id, addr := range walletAddrs {
		log.Infof("wallet %v is: %v", id, addr)
	}

	err = ConnectAuthor(ctx, params)
	if err != nil {
		return nil, err
	}

	adminTokenV, err := params.VenusAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	err = GetWalletInfo(ctx, params, adminTokenV.MustString(), walletAddrs[0])
	if err != nil {
		log.Infof("get wallet info failed: %v", err)
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
		return nil, fmt.Errorf("exec remote cmd failed: %w", err)
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
		return fmt.Errorf("exec remote cmd failed: %w", err)
	}

	return nil
}

func GetWalletInfo(ctx context.Context, params TestCaseParams, authToken string, walletAddr string) error {
	endpoint, err := params.VenusAuth.SvcEndpoint()
	if err != nil {
		return err
	}
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

	api, closer, err := v2API.DialIGatewayRPC(ctx, endpoint.ToHTTP(), authToken, nil)
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
