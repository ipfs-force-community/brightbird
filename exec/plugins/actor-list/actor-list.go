package main

import (
	"bytes"
	"context"
	"fmt"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
	"text/tabwriter"
)

var Info = types.PluginInfo{
	Name:        "actor-list",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "actor list",
}

type TestCaseParams struct {
	fx.In
	AdminToken  types.AdminToken
	K8sEnv      *env.K8sEnvDeployer      `json:"-"`
	VenusMarket env.IVenusMarketDeployer `json:"-"`
}

func Exec(ctx context.Context, params TestCaseParams) error {

	listenAddress, err := actorList(ctx, params)
	if err != nil {
		fmt.Printf("market net listen err: %v\n", err)
		return err
	}
	fmt.Printf("market net listen is: %v\n", listenAddress)

	return nil
}

func actorList(ctx context.Context, params TestCaseParams) (string, error) {
	endpoint := params.VenusMarket.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, params.VenusMarket.Pods()[0].GetName(), int(params.VenusMarket.Svc().Spec.Ports[0].Port))
		if err != nil {
			return "", err
		}
	}
	client, closer, err := marketapi.NewIMarketRPC(ctx, endpoint.ToHttp(), nil)
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
