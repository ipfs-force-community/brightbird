package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/docker/go-units"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	venusAPI "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletmarket "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("retrieval-deals-set-ask")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "retrieval-deals-set-ask",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "retrieval deals set-ask",
}

type TestCaseParams struct {
	Auth           sophonauth.SophonAuthDeployReturn       `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	DamoclesMarket dropletmarket.DropletMarketDeployReturn `json:"DropletMarket" description:"droplet market return"`
	Venus          venus.VenusDeployReturn                 `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	MinerAddress   address.Address                         `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true" `
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	fullNode, closer, err := venusAPI.DialFullNodeRPC(ctx, params.Venus.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	minerInfo, err := fullNode.StateMinerInfo(ctx, params.MinerAddress, vTypes.EmptyTSK)
	if err != nil {
		fmt.Printf("get miner info failed: %v\n", err)
		return err
	}
	log.Infof("miner info: %v", minerInfo)

	err = StorageAskSet(ctx, params, params.MinerAddress)
	if err != nil {
		fmt.Printf("market net listen err: %v\n", err)
		return err
	}

	err = StorageGetAsk(ctx, params, params.MinerAddress)
	if err != nil {
		fmt.Printf("market net listen err: %v\n", err)
		return err
	}

	return nil
}

func StorageAskSet(ctx context.Context, params TestCaseParams, mAddr address.Address) error {
	client, closer, err := marketapi.NewIMarketRPC(ctx, params.DamoclesMarket.SvcEndpoint.ToMultiAddr(), nil)
	if err != nil {
		return err
	}
	defer closer()

	ask, err := client.MarketGetRetrievalAsk(ctx, mAddr)
	if err != nil {
		if !strings.Contains(err.Error(), "record not found") {
			return err
		}
		ask = &retrievalmarket.Ask{}
	}

	priceStr := "0.0000001"
	price, err := vTypes.ParseFIL(priceStr)
	if err != nil {
		return err
	}
	ask.PricePerByte = vTypes.BigDiv(vTypes.BigInt(price), vTypes.NewInt(1<<30))

	unsealPriceStr := "0.0000001"
	unsealPrice, err := vTypes.ParseFIL(unsealPriceStr)
	if err != nil {
		return err
	}
	ask.UnsealPrice = abi.TokenAmount(unsealPrice)

	paymentIntervalStr := "100MB"
	paymentInterval, err := units.RAMInBytes(paymentIntervalStr)
	if err != nil {
		return err
	}
	ask.PaymentInterval = uint64(paymentInterval)

	paymentIntervalIncreaseStr := "100"
	paymentIntervalIncrease, err := units.RAMInBytes(paymentIntervalIncreaseStr)
	if err != nil {
		return err
	}
	ask.PaymentIntervalIncrease = uint64(paymentIntervalIncrease)

	err = client.MarketSetRetrievalAsk(ctx, mAddr, ask)
	if err != nil {
		return err
	}

	return err
}

func StorageGetAsk(ctx context.Context, params TestCaseParams, mAddr address.Address) error {
	client, closer, err := marketapi.NewIMarketRPC(ctx, params.DamoclesMarket.SvcEndpoint.ToMultiAddr(), nil)
	if err != nil {
		return err
	}
	defer closer()

	sask, err := client.MarketGetAsk(ctx, mAddr)
	if err != nil {
		return err
	}

	var ask *storagemarket.StorageAsk
	if sask != nil && sask.Ask != nil {
		ask = sask.Ask
	}

	w := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
	fmt.Fprintf(w, "Price per GiB/Epoch\tVerified\tMin. Piece Size (padded)\tMax. Piece Size (padded)\tExpiry (Epoch)\tExpiry (Appx. Rem. Time)\tSeq. No.\n")
	if ask == nil {
		fmt.Fprintf(w, "<miner does not have an ask>\n")
		return w.Flush()
	}
	return nil
}
