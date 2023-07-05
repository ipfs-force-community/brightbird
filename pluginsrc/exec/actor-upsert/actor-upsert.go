package main

import (
	"bytes"
	"context"
	"fmt"
	"text/tabwriter"

	dropletmarket "github.com/hunjixin/brightbird/pluginsrc/deploy/droplet-market"
	"github.com/hunjixin/brightbird/types"

	"github.com/filecoin-project/go-address"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	mkTypes "github.com/filecoin-project/venus/venus-shared/types/market"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/version"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "actor-upsert",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "actor upsert",
}

type TestCaseParams struct {
	DropletMarket dropletmarket.DropletMarketDeployReturn `json:"DropletMarket" description:"droplet market return "`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	mAddr, err := actorUpsert(ctx, params)
	if err != nil {
		return fmt.Errorf("upsert actor err %w", err)
	}

	id, err := actorList(ctx, params, mAddr)
	if id == "" {
		return fmt.Errorf("list actor err %w", err)
	}

	return nil
}

func actorUpsert(ctx context.Context, params TestCaseParams) (address.Address, error) {
	client, closer, err := marketapi.NewIMarketRPC(ctx, params.DropletMarket.SvcEndpoint.ToMultiAddr(), nil)
	if err != nil {
		return address.Undef, err
	}
	defer closer()

	miner := "t01999"
	mAddr, err := address.NewFromString(miner)
	if err != nil {
		return address.Undef, err
	}

	bAdd, err := client.ActorUpsert(ctx, mkTypes.User{Addr: mAddr})
	if err != nil {
		return address.Undef, nil
	}

	opr := "Add"
	if !bAdd {
		opr = "Update"
	}

	fmt.Printf("%s miner %s success\n", opr, mAddr)

	return mAddr, err
}

func actorList(ctx context.Context, params TestCaseParams, mAddr address.Address) (string, error) {
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
		if miner.Addr == mAddr {
			return miner.Addr.String(), nil
		}
	}

	return "", err
}
