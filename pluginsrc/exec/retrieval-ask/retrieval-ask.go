package main

import (
	"context"
	"strings"

	"github.com/docker/go-units"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-state-types/abi"
	venusAPI "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletmarket "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("retrieval-ask")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "retrieval-ask",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "配置（设置/更新）提供商的 retrieval ask，获取提供商当前的 retrieval ask",
}

type TestCaseParams struct {
	Auth    sophonauth.SophonAuthDeployReturn       `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Droplet dropletmarket.DropletMarketDeployReturn `json:"Droplet" jsonschema:"Droplet" title:"Droplet" require:"true" description:"droplet return"`
	Venus   venus.VenusDeployReturn                 `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`

	MinerAddress            address.Address `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true"`
	Price                   vTypes.FIL      `json:"price"  jsonschema:"price"  title:"price" default:"0.0fil" require:"true" description:"Set the price of the ask for retrievals"`
	UnsealPrice             vTypes.FIL      `json:"unsealPrice"  jsonschema:"unsealPrice"  title:"unsealPrice" default:"0.0fil" require:"true" description:"Set the price to unseal"`
	PaymentInterval         string          `json:"paymentInterval"  jsonschema:"paymentInterval"  title:"paymentInterval" default:"1MiB" require:"true" description:"Set the payment interval (in bytes) for retrieval"`
	PaymentIntervalIncrease string          `json:"paymentIntervalIncrease"  jsonschema:"paymentIntervalIncrease"  title:"paymentIntervalIncrease" default:"1MiB" require:"true" description:"Set the payment interval increase (in bytes) for retrieval"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	fullNode, closer, err := venusAPI.DialFullNodeRPC(ctx, params.Venus.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	client, closer, err := marketapi.DialIMarketRPC(ctx, params.Droplet.SvcEndpoint.ToMultiAddr(), params.Droplet.UserToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	minerInfo, err := fullNode.StateMinerInfo(ctx, params.MinerAddress, vTypes.EmptyTSK)
	if err != nil {
		log.Errorln("get miner info failed: %v\n", err)
		return err
	}
	log.Infof("miner info: %v", minerInfo)

	err = StorageAskSet(ctx, params, params.MinerAddress, client)
	if err != nil {
		log.Errorln("market net listen err: %v\n", err)
		return err
	}

	err = StorageGetAsk(ctx, params, params.MinerAddress, client)
	if err != nil {
		log.Errorln("market net listen err: %v\n", err)
		return err
	}

	return nil
}

func StorageAskSet(ctx context.Context, params TestCaseParams, mAddr address.Address, client marketapi.IMarket) error {
	ask, err := client.MarketGetRetrievalAsk(ctx, mAddr)
	if err != nil {
		if !strings.Contains(err.Error(), "record not found") {
			return err
		}
		ask = &retrievalmarket.Ask{}
	}

	ask.PricePerByte = vTypes.BigDiv(vTypes.BigInt(params.Price), vTypes.NewInt(1<<30))
	ask.UnsealPrice = abi.TokenAmount(params.UnsealPrice)

	paymentInterval, _ := units.RAMInBytes(params.PaymentInterval)
	ask.PaymentInterval = uint64(paymentInterval)

	paymentIntervalIncrease, _ := units.RAMInBytes(params.PaymentIntervalIncrease)
	ask.PaymentIntervalIncrease = uint64(paymentIntervalIncrease)

	err = client.MarketSetRetrievalAsk(ctx, mAddr, ask)
	if err != nil {
		return err
	}

	return err
}

func StorageGetAsk(ctx context.Context, params TestCaseParams, mAddr address.Address, client marketapi.IMarket) error {
	ask, err := client.MarketGetRetrievalAsk(ctx, mAddr)
	if err != nil {
		if !strings.Contains(err.Error(), "record not found") {
			return err
		}
		return err
	}
	log.Debugln(ask)
	return nil
}
