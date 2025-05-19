package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	v2API "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"

	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	sophongateway "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-gateway"
)

var log = logging.Logger(Info.Name)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "get-wallet-in-gateway",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "search wallet from sophon gateway",
}

type TestCaseParams struct {
	Auth          sophonauth.SophonAuthDeployReturn `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Gateway       sophongateway.SophonGatewayReturn `json:"SophonGateway"  jsonschema:"SophonGateway"  title:"SophonGateway" require:"true" description:"gateway deploy return"`
	WalletAddress address.Address                   `json:"walletAddress" jsonschema:"walletAddress" title:"walletAddress" require:"false" description:"walletAddress"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	api, closer, err := v2API.DialIGatewayRPC(ctx, params.Gateway.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	// log.Debugln("param wallet is", params.WalletAddress)

	wallets, err := api.ListWalletInfo(ctx)
	if err != nil {
		return err
	}
	isFound := false
	for i, wallet := range wallets {
		log.Debugf("wallet list [%v] : \n%v", i, wallet)
		for _, addr := range wallet.ConnectStates {
			for _, a := range addr.Addrs {
				if a == params.WalletAddress {
					isFound = true
				}
			}
		}
	}

	if !isFound {
		return fmt.Errorf("wallet %s not found in gateway", params.WalletAddress)
	}

	log.Infof("wallet %s found in gateway", params.WalletAddress)
	return nil
}
