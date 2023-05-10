package main

import (
	"bytes"
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/hunjixin/brightbird/types"

	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
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
	fx.In
	K8sEnv      *env.K8sEnvDeployer `json:"-"`
	VenusMarket env.IDeployer       `json:"-" svcname:"VenusWallet"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {

	listenAddress, err := actorList(ctx, params)
	if err != nil {
		fmt.Printf("market net listen err: %v\n", err)
		return nil, err
	}
	fmt.Printf("market net listen is: %v\n", listenAddress)

	return env.NewSimpleExec(), nil
}

func actorList(ctx context.Context, params TestCaseParams) (string, error) {
	endpoint, err := params.VenusMarket.SvcEndpoint()
	if err != nil {
		return "", err
	}
	if env.Debug {
		pods, err := params.VenusMarket.Pods(ctx)
		if err != nil {
			return "", err
		}

		svc, err := params.VenusMarket.Svc(ctx)
		if err != nil {
			return "", err
		}
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
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
