package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/docker/go-units"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/pkg/constants"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/filecoin-project/venus/venus-shared/types/market"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "storage-deals-set-ask",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "storage deals set-ask",
}

type TestCaseParams struct {
	fx.In
	K8sEnv                     *env.K8sEnvDeployer `json:"-"`
	VenusMarket                env.IDeployer       `json:"-" svcname:"VenusMarket"`
	VenusSectorManagerDeployer env.IDeployer       `json:"-" svcname:"VenusSectorManager"`
	CreateWallet               env.IExec           `json:"-" svcname:"CreateWallet"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {

	walletAddr, err := params.CreateWallet.Param("CreateWallet")
	if err != nil {
		return nil, err
	}

	minerAddr, err := CreateMiner(ctx, params, walletAddr.(address.Address))
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

	err = StorageAskGet(ctx, params, minerAddr)
	if err != nil {
		fmt.Printf("market net listen err: %v\n", err)
		return nil, err
	}

	return env.NewSimpleExec(), nil
}

func StorageAskSet(ctx context.Context, params TestCaseParams, mAddr address.Address) error {
	endpoint := params.VenusMarket.SvcEndpoint()
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

	isUpdate := true
	storageAsk, err := client.MarketGetAsk(ctx, mAddr)
	if err != nil {
		if !strings.Contains(err.Error(), "record not found") {
			return err
		}
		storageAsk = &market.SignedStorageAsk{}
		isUpdate = false
	}

	pri, err := vTypes.ParseFIL("0.000000001")
	if err != nil {
		return err
	}

	vpri, err := vTypes.ParseFIL("0")
	if err != nil {
		return err
	}

	dur, err := time.ParseDuration("720h0m0s")
	if err != nil {
		return fmt.Errorf("cannot parse duration: %w", err)
	}

	qty := dur.Seconds() / float64(constants.MainNetBlockDelaySecs)

	min, err := units.RAMInBytes("1KB")
	if err != nil {
		return fmt.Errorf("cannot parse min-piece-size to quantity of bytes: %w", err)
	}

	if min < 256 {
		return errors.New("minimum piece size (w/bit-padding) is 256B")
	}

	max, err := units.RAMInBytes("32GiB")
	if err != nil {
		return fmt.Errorf("cannot parse max-piece-size to quantity of bytes: %w", err)
	}

	ssize, err := client.ActorSectorSize(ctx, mAddr)
	if err != nil {
		return fmt.Errorf("get miner's size %w", err)
	}

	smax := int64(ssize)

	if max == 0 {
		max = smax
	}

	if max > smax {
		return fmt.Errorf("max piece size (w/bit-padding) %s cannot exceed miner sector size %s", vTypes.SizeStr(vTypes.NewInt(uint64(max))), vTypes.SizeStr(vTypes.NewInt(uint64(smax))))
	}

	if isUpdate {
		storageAsk.Ask.Price = vTypes.BigInt(pri)
		storageAsk.Ask.VerifiedPrice = vTypes.BigInt(vpri)
		storageAsk.Ask.MinPieceSize = abi.PaddedPieceSize(min)
		storageAsk.Ask.MaxPieceSize = abi.PaddedPieceSize(max)
		return client.MarketSetAsk(ctx, mAddr, storageAsk.Ask.Price, storageAsk.Ask.VerifiedPrice, abi.ChainEpoch(qty), storageAsk.Ask.MinPieceSize, storageAsk.Ask.MaxPieceSize)
	}

	err = client.MarketSetAsk(ctx, mAddr, vTypes.BigInt(pri), vTypes.BigInt(vpri), abi.ChainEpoch(qty), abi.PaddedPieceSize(min), abi.PaddedPieceSize(max))
	if err != nil {
		return fmt.Errorf("market set ask err %w", err)
	}

	return err
}

func StorageAskGet(ctx context.Context, params TestCaseParams, mAddr address.Address) error {
	endpoint := params.VenusMarket.SvcEndpoint()
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
