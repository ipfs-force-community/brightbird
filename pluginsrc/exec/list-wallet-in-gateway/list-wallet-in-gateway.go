package main

import (
	"context"

	"github.com/filecoin-project/go-address"
	v2API "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	gtypes "github.com/filecoin-project/venus/venus-shared/types/gateway"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	sophongateway "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-gateway"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("list-wallet-in-gateway")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "list-wallet-in-gateway",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "从gateway获取wallet信息",
}

type TestCaseParams struct {
	Auth          sophonauth.SophonAuthDeployReturn `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Gateway       sophongateway.SophonGatewayReturn `json:"SophonGateway"  jsonschema:"SophonGateway"  title:"SophonGateway" require:"true" description:"gateway deploy return"`
	WalletAddress address.Address                   `json:"walletAddress" jsonschema:"walletAddress" title:"walletAddress" require:"false" description:"walletAddress"`
}

// todo support array
type ListWalletInGatewayReturn = []*gtypes.WalletDetail

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (ListWalletInGatewayReturn, error) {
	api, closer, err := v2API.DialIGatewayRPC(ctx, params.Gateway.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	log.Debugln("param wallet is", params.WalletAddress)

	wallets, err := api.ListWalletInfo(ctx)
	if err != nil {
		return nil, err
	}
	for i, wallet := range wallets {
		for _, addr := range wallet.ConnectStates {
			log.Debugf("wallet list %v is %v\n", i, addr.Addrs)
		}
	}

	return wallets, nil
}
