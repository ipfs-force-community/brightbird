package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"go.uber.org/fx"

	"github.com/docker/go-units"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
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
	fx.In
	K8sEnv          *env.K8sEnvDeployer `json:"-"`
	DamoclesMarket  env.IDeployer       `json:"-" svcname:"DamoclesMarket"`
	DamoclesManager env.IDeployer       `json:"-" svcname:"DamoclesManager"`
	CreateMiner     env.IExec           `json:"-" svcname:"CreateMiner"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {

	minerAddr, err := params.CreateMiner.Param("Miner")
	if err != nil {
		return nil, err
	}

	addr, err := env.UnmarshalJSON[address.Address](minerAddr.Raw())
	if err != nil {
		return nil, err
	}

	minerInfo, err := GetMinerInfo(ctx, params, addr)
	if err != nil {
		fmt.Printf("get miner info failed: %v\n", err)
		return nil, err
	}
	log.Infof("miner info: %v", minerInfo)

	err = StorageAskSet(ctx, params, addr)
	if err != nil {
		fmt.Printf("market net listen err: %v\n", err)
		return nil, err
	}

	err = StorageGetAsk(ctx, params, addr)
	if err != nil {
		fmt.Printf("market net listen err: %v\n", err)
		return nil, err
	}

	return env.NewSimpleExec(), nil
}

func StorageAskSet(ctx context.Context, params TestCaseParams, mAddr address.Address) error {
	endpoint, err := params.DamoclesMarket.SvcEndpoint()
	if err != nil {
		return err
	}

	client, closer, err := marketapi.NewIMarketRPC(ctx, endpoint.ToHTTP(), nil)
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
	endpoint, err := params.DamoclesMarket.SvcEndpoint()
	if err != nil {
		return err
	}

	client, closer, err := marketapi.NewIMarketRPC(ctx, endpoint.ToHTTP(), nil)
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

func GetMinerInfo(ctx context.Context, params TestCaseParams, minerAddr address.Address) (string, error) {
	getMinerCmd := []string{
		"./venus-sector-manager",
		"util",
		"miner",
		"info",
		minerAddr.String(),
	}

	pods, err := params.DamoclesManager.Pods(ctx)
	if err != nil {
		return "", err
	}

	minerInfo, err := params.K8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), getMinerCmd...)
	if err != nil {
		return "", fmt.Errorf("exec remote cmd failed: %w", err)
	}

	return string(minerInfo), nil
}
