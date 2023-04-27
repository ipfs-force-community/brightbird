package main

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-address"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	mkTypes "github.com/filecoin-project/venus/venus-shared/types/market"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "actor-upsert",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "actor upsert",
}

type TestCaseParams struct {
	fx.In
	AdminToken  types.AdminToken
	K8sEnv      *env.K8sEnvDeployer      `json:"-"`
	VenusMarket env.IVenusMarketDeployer `json:"-"`
}

func Exec(ctx context.Context, params TestCaseParams) error {

	err := actorUpsert(ctx, params)
	if err != nil {
		fmt.Printf("market net listen err: %v\n", err)
		return err
	}

	return nil
}

func actorUpsert(ctx context.Context, params TestCaseParams) error {
	endpoint := params.VenusMarket.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, params.VenusMarket.Pods()[0].GetName(), int(params.VenusMarket.Svc().Spec.Ports[0].Port))
		if err != nil {
			return err
		}
	}
	client, closer, err := marketapi.NewIMarketRPC(ctx, endpoint.ToHttp(), nil)
	if err != nil {
		return err
	}
	defer closer()

	miner := "t01999"
	mAddr, err := address.NewFromString(miner)
	if err != nil {
		return err
	}

	bAdd, err := client.ActorUpsert(ctx, mkTypes.User{Addr: mAddr})
	if err != nil {
		return nil
	}

	opr := "Add"
	if !bAdd {
		opr = "Update"
	}

	fmt.Printf("%s miner %s success\n", opr, mAddr)

	return err
}
