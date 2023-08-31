package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"

	v2API "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	sophongateway "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-gateway"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "get-miner-from-gateway",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "从gateway检查能否获取到期望的miner",
}

type TestCaseParams struct {
	Auth    sophonauth.SophonAuthDeployReturn `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Gateway sophongateway.SophonGatewayReturn `json:"SophonGateway"  jsonschema:"SophonGateway"  title:"SophonGateway" require:"true" description:"gateway deploy return"`
	Miner   address.Address                   `json:"miner" jsonschema:"miner" title:"Miner Address" require:"true" description:"miner address"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	api, closer, err := v2API.DialIGatewayRPC(ctx, params.Gateway.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	minerList, err := api.ListConnectedMiners(ctx)
	if err != nil {
		return err
	}

	fmt.Println("全部的miner列表:")
	for _, miner := range minerList {
		fmt.Println(miner)
		if miner == params.Miner {
			return nil
		}
	}
	return fmt.Errorf("miner %s not found", params.Miner)
}
