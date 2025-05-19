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
	Name:        "get-miner-from-gateway",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "search miner from gateway",
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

	log.Debugln("miner list :")
	isFound := false
	for _, miner := range minerList {
		log.Debugf("miner: %s", miner.String())
		if miner == params.Miner {
			isFound = true
		}
	}
	if !isFound {
		return fmt.Errorf("miner %s not found in gateway", params.Miner)
	} else {
		log.Infof("miner %s found in gateway", params.Miner)
	}

	return nil
}
