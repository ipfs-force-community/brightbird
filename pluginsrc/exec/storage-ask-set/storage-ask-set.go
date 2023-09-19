package main

import (
	"context"
	"strings"
	"time"

	"github.com/docker/go-units"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/pkg/constants"
	venusAPI "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/filecoin-project/venus/venus-shared/types/market"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletmarket "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("storage-ask-set")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "storage-ask-set",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "配置（设置/更新）stroage ask",
}

type TestCaseParams struct {
	Auth    sophonauth.SophonAuthDeployReturn       `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Venus   venus.VenusDeployReturn                 `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	Droplet dropletmarket.DropletMarketDeployReturn `json:"Droplet" jsonschema:"Droplet" title:"Droplet" description:"droplet return"`

	MinerAddress address.Address `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true" `

	Price         vTypes.FIL `json:"price"  jsonschema:"price"  title:"price" default:"0.01fil" require:"true" description:"price(fil)"`
	VerifiedPrice vTypes.FIL `json:"verifiedPrice"  jsonschema:"verifiedPrice"  title:"verifiedPrice" default:"0.02fil" require:"true" description:"verified price(fil)"`
	MinPriceSize  string     `json:"minPriceSize"  jsonschema:"minPriceSize"  title:"minPriceSize" default:"512b" require:"true" description:"size"`
	MaxPriceSize  string     `json:"maxPriceSize"  jsonschema:"maxPriceSize"  title:"maxPriceSize" require:"true" description:"size"`
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
		return err
	}

	return nil
}

func StorageAskSet(ctx context.Context, params TestCaseParams, mAddr address.Address, client marketapi.IMarket) error {
	isUpdate := true
	storageAsk, err := client.MarketGetAsk(ctx, mAddr)
	if err != nil {
		if !strings.Contains(err.Error(), "record not found") {
			return err
		}
		storageAsk = &market.SignedStorageAsk{}
		isUpdate = false
	}

	dur, err := time.ParseDuration("720h0m0s")
	if err != nil {
		log.Errorln("cannot parse duration: %w", err)
		return err
	}

	qty := dur.Seconds() / float64(constants.MainNetBlockDelaySecs)

	min, _ := units.RAMInBytes(params.MinPriceSize)
	var max int64
	if params.MaxPriceSize != "" {
		max, _ = units.RAMInBytes(params.MaxPriceSize)
	}
	ssize, err := client.ActorSectorSize(ctx, mAddr)
	if err != nil {
		log.Errorln("get miner's size %w", err)
		return err
	}

	smax := int64(ssize)
	if max == 0 {
		max = smax
	}
	if max > smax {
		log.Errorln("max piece size (w/bit-padding) %s cannot exceed miner sector size %s", vTypes.SizeStr(vTypes.NewInt(uint64(max))), vTypes.SizeStr(vTypes.NewInt(uint64(smax))))
		return err
	}

	if isUpdate {
		storageAsk.Ask.Price = vTypes.BigInt(params.Price)
		storageAsk.Ask.VerifiedPrice = vTypes.BigInt(params.VerifiedPrice)
		storageAsk.Ask.MinPieceSize = abi.PaddedPieceSize(min)
		storageAsk.Ask.MaxPieceSize = abi.PaddedPieceSize(max)
		return client.MarketSetAsk(ctx, mAddr, storageAsk.Ask.Price, storageAsk.Ask.VerifiedPrice, abi.ChainEpoch(qty), storageAsk.Ask.MinPieceSize, storageAsk.Ask.MaxPieceSize)
	}

	err = client.MarketSetAsk(ctx, mAddr, vTypes.BigInt(params.Price), vTypes.BigInt(params.VerifiedPrice), abi.ChainEpoch(qty), abi.PaddedPieceSize(min), abi.PaddedPieceSize(max))
	if err != nil {
		log.Errorln("market set ask err %w", err)
		return err
	}

	return nil
}
