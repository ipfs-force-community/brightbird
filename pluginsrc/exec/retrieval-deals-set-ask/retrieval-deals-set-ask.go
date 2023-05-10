package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/docker/go-units"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	"github.com/filecoin-project/venus/venus-shared/api/wallet"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

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
	K8sEnv                     *env.K8sEnvDeployer `json:"-"`
	VenusMarket                env.IDeployer       `json:"-" svcname:"VenusMessage"`
	VenusWallet                env.IDeployer       `json:"-" svcname:"VenusWallet"`
	VenusSectorManagerDeployer env.IDeployer       `json:"-" svcname:"VenusMessage"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {

	walletAddr, err := CreateWallet(ctx, params)
	if err != nil {
		fmt.Printf("create wallet failed: %v\n", err)
		return nil, err
	}

	minerAddr, err := CreateMiner(ctx, params, walletAddr)
	if err != nil {
		fmt.Printf("create miner failed: %v\n", err)
		return nil, err
	}

	minerInfo, err := GetMinerInfo(ctx, params, minerAddr)
	if err != nil {
		fmt.Printf("get miner info failed: %v\n", err)
		return nil, err
	}
	fmt.Println("miner info: %v", minerInfo)

	err = StorageAskSet(ctx, params, minerAddr)
	if err != nil {
		fmt.Printf("market net listen err: %v\n", err)
		return nil, err
	}

	err = StorageGetAsk(ctx, params, minerAddr)
	if err != nil {
		fmt.Printf("market net listen err: %v\n", err)
		return nil, err
	}

	return env.NewSimpleExec(), nil
}

func StorageAskSet(ctx context.Context, params TestCaseParams, mAddr address.Address) error {
	endpoint, err := params.VenusMarket.SvcEndpoint()
	if err != nil {
		return err
	}

	if env.Debug {
		pods, err := params.VenusMarket.Pods(ctx)
		if err != nil {
			return err
		}

		svc, err := params.VenusMarket.Svc(ctx)
		if err != nil {
			return err
		}
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return err
		}
	}
	client, closer, err := marketapi.NewIMarketRPC(ctx, endpoint.ToHttp(), nil)
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

	price_str := "0.0000001"
	price, err := vTypes.ParseFIL(price_str)
	if err != nil {
		return err
	}
	ask.PricePerByte = vTypes.BigDiv(vTypes.BigInt(price), vTypes.NewInt(1<<30))

	unseal_price := "0.0000001"
	unsealPrice, err := vTypes.ParseFIL(unseal_price)
	if err != nil {
		return err
	}
	ask.UnsealPrice = abi.TokenAmount(unsealPrice)

	payment_interval := "100MB"
	paymentInterval, err := units.RAMInBytes(payment_interval)
	if err != nil {
		return err
	}
	ask.PaymentInterval = uint64(paymentInterval)

	payment_interval_increase := "100"
	paymentIntervalIncrease, err := units.RAMInBytes(payment_interval_increase)
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
	endpoint, err := params.VenusMarket.SvcEndpoint()
	if err != nil {
		return err
	}

	if env.Debug {
		pods, err := params.VenusMarket.Pods(ctx)
		if err != nil {
			return err
		}

		svc, err := params.VenusMarket.Svc(ctx)
		if err != nil {
			return err
		}
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return err
		}
	}
	client, closer, err := marketapi.NewIMarketRPC(ctx, endpoint.ToHttp(), nil)
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

func CreateWallet(ctx context.Context, params TestCaseParams) (address.Address, error) {
	pods, err := params.VenusWallet.Pods(ctx)
	if err != nil {
		return address.Undef, err
	}

	svc, err := params.VenusWallet.Svc(ctx)
	if err != nil {
		return address.Undef, err
	}

	walletToken, err := env.ReadWalletToken(ctx, params.K8sEnv, pods[0].GetName())
	if err != nil {
		return address.Undef, fmt.Errorf("read wallet token failed: %w\n", err)
	}

	endpoint, err := params.VenusWallet.SvcEndpoint()
	if err != nil {
		return address.Undef, err
	}

	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return address.Undef, fmt.Errorf("port forward failed: %w\n", err)
		}
	}

	walletRpc, closer, err := wallet.DialIFullAPIRPC(ctx, endpoint.ToMultiAddr(), walletToken, nil)
	if err != nil {
		return address.Undef, fmt.Errorf("dial iFullAPI rpc failed: %w\n", err)
	}
	defer closer()

	password := "123456"
	err = walletRpc.SetPassword(ctx, password)
	if err != nil {
		return address.Undef, fmt.Errorf("set password failed: %w\n", err)
	}

	walletAddr, err := walletRpc.WalletNew(ctx, vTypes.KTBLS)
	if err != nil {
		return address.Undef, fmt.Errorf("create wallet failed: %w\n", err)
	}
	fmt.Printf("wallet: %v\n", walletAddr)

	return walletAddr, nil
}

func CreateMiner(ctx context.Context, params TestCaseParams, walletAddr address.Address) (address.Address, error) {
	cmd := []string{
		"./venus-sector-manager",
		"util",
		"miner",
		"create",
		"--sector-size=8MiB",
		"--exid=" + string(rune(rand.Intn(100000))),
	}
	cmd = append(cmd, "--from="+walletAddr.String())

	pods, err := params.VenusSectorManagerDeployer.Pods(ctx)
	if err != nil {
		return address.Undef, err
	}

	minerAddr, err := params.K8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), cmd...)
	if err != nil {
		return address.Undef, fmt.Errorf("exec remote cmd failed: %w\n", err)
	}

	addr, err := address.NewFromBytes(minerAddr)
	if err != nil {
		return address.Undef, err
	}
	return addr, nil
}

func GetMinerInfo(ctx context.Context, params TestCaseParams, minerAddr address.Address) (string, error) {
	getMinerCmd := []string{
		"./venus-sector-manager",
		"util",
		"miner",
		"info",
		minerAddr.String(),
	}

	pods, err := params.VenusSectorManagerDeployer.Pods(ctx)
	if err != nil {
		return "", err
	}

	minerInfo, err := params.K8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), getMinerCmd...)
	if err != nil {
		return "", fmt.Errorf("exec remote cmd failed: %w\n", err)
	}

	return string(minerInfo), nil
}
