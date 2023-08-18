package main

import (
	"bytes"
	"context"
	"fmt"
	"text/tabwriter"

	dropletmarket "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
	"github.com/ipfs-force-community/brightbird/types"

	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	"github.com/ipfs-force-community/brightbird/version"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "actor-list",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "actor list",
}

type TestCaseParams struct {
	DropletMarket dropletmarket.DropletMarketDeployReturn `json:"DropletMarket" jsonschema:"DropletMarket" title:"DropletMarket" description:"droplet market return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	listenAddress, err := actorList(ctx, params)
	if err != nil {
		return fmt.Errorf("list actor err:%w", err)
	}
	fmt.Printf("market net listen is: %v\n", listenAddress)
	return nil
}

func actorList(ctx context.Context, params TestCaseParams) (string, error) {
	client, closer, err := marketapi.NewIMarketRPC(ctx, params.DropletMarket.SvcEndpoint.ToMultiAddr(), nil)
	if err != nil {
		return "", err
	}
	defer closer()

	miners, err := client.ActorList(ctx)
	if err != nil {
		return "", nil
	}

	buf := &bytes.Buffer{}
	tw := tabwriter.NewWriter(buf, 2, 4, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "miner\taccount")
	for _, miner := range miners {
		_, _ = fmt.Fprintf(tw, "%s\t%s\n", miner.Addr.String(), miner.Account)
	}
	if err := tw.Flush(); err != nil {
		return "", err
	}
	fmt.Println(buf.String())

	return buf.String(), err
}
