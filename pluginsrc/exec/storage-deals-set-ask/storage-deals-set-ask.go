package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/docker/go-units"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/pkg/constants"
	venusAPI "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/filecoin-project/venus/venus-shared/types/market"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	damoclesmanager "github.com/hunjixin/brightbird/pluginsrc/deploy/damocles-manager"
	dropletmarket "github.com/hunjixin/brightbird/pluginsrc/deploy/droplet-market"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/hunjixin/brightbird/pluginsrc/deploy/venus"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("storage-deals")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "storage-deals-set-ask",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "storage deals set-ask",
}

type TestCaseParams struct {
	SophonAuth      sophonauth.SophonAuthDeployReturn       `json:"SophonAuth"`
	Venus           venus.VenusDeployReturn                 `json:"Venus" description:"venus return"`
	DropletMarket   dropletmarket.DropletMarketDeployReturn `json:"DropletMarket"`
	DamoclesManager damoclesmanager.DamoclesManagerReturn   `json:"DamoclesManager"`
	MinerAddress    address.Address                         `json:"minerAddress" type:"string"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	fullNode, closer, err := venusAPI.DialFullNodeRPC(ctx, params.Venus.SvcEndpoint.ToMultiAddr(), params.SophonAuth.AdminToken, nil)
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
		return err
	}

	err = StorageAskGet(ctx, params, params.MinerAddress)
	if err != nil {
		return err
	}

	return nil
}

func StorageAskSet(ctx context.Context, params TestCaseParams, mAddr address.Address) error {
	client, closer, err := marketapi.NewIMarketRPC(ctx, params.DropletMarket.SvcEndpoint.ToMultiAddr(), nil)
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
	client, closer, err := marketapi.NewIMarketRPC(ctx, params.DropletMarket.SvcEndpoint.ToMultiAddr(), nil)
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
