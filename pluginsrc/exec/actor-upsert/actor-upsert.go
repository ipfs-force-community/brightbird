package main

import (
	"bytes"
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/hunjixin/brightbird/types"

	"github.com/filecoin-project/go-address"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	mkTypes "github.com/filecoin-project/venus/venus-shared/types/market"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
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
	fx.In

	K8sEnv      *env.K8sEnvDeployer `json:"-"`
	VenusMarket env.IDeployer       `json:"-" svcname:"VenusMarket"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {

	mAddr, err := actorUpsert(ctx, params)
	if err != nil {
		fmt.Printf("market net listen err: %v\n", err)
		return nil, err
	}

	id, err := actorList(ctx, params, mAddr)
	if id == "" {
		fmt.Printf("actor delete err: %v\n", err)
		return nil, err
	}

	return env.NewSimpleExec(), nil
}

func actorUpsert(ctx context.Context, params TestCaseParams) (address.Address, error) {
	endpoint, err := params.VenusMarket.SvcEndpoint()
	if err != nil {
		return address.Undef, err
	}
	if env.Debug {
		pods, err := params.VenusMarket.Pods(ctx)
		if err != nil {
			return address.Undef, err
		}

		svc, err := params.VenusMarket.Svc(ctx)
		if err != nil {
			return address.Undef, err
		}

		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return address.Undef, err
		}
	}
	client, closer, err := marketapi.NewIMarketRPC(ctx, endpoint.ToHTTP(), nil)
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
	client, closer, err := marketapi.NewIMarketRPC(ctx, endpoint.ToHTTP(), nil)
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
