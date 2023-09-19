package main

import (
	"context"
	"sort"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	venusAPI "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	clientapi "github.com/filecoin-project/venus/venus-shared/api/market/client"
	vtypes "github.com/filecoin-project/venus/venus-shared/types"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletclient "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-client"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("clinet-storage-asks-list")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "clinet-storage-asks-list",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "droplet-client 查询 miner 列表",
}

type TestCaseParams struct {
	Venus         venus.VenusDeployReturn                 `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	DropletClient dropletclient.DropletClientDeployReturn `json:"DropClient" jsonschema:"DropClient" title:"DropletClient" description:"droplet client return"`
	MinerAddress  address.Address                         `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true" `
}

type StorageAsksListReturn []*storagemarket.StorageAsk

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (StorageAsksListReturn, error) {
	api, closer, err := clientapi.DialIMarketClientRPC(ctx, params.DropletClient.SvcEndpoint.ToMultiAddr(), params.DropletClient.ClientToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	fapi, closer, err := venusAPI.DialFullNodeRPC(ctx, params.Venus.SvcEndpoint.ToMultiAddr(), params.Venus.AdminToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	asks, err := GetAsks(ctx, fapi, api)
	if err != nil {
		return nil, err
	}

	pfmt := "%s: min:%s max:%s price:%s/GiB/Epoch verifiedPrice:%s\n"
	for _, ask := range asks {
		log.Debugf(pfmt, ask.Miner,
			vtypes.SizeStr(vtypes.NewInt(uint64(ask.MinPieceSize))),
			vtypes.SizeStr(vtypes.NewInt(uint64(ask.MaxPieceSize))),
			vtypes.FIL(ask.Price),
			vtypes.FIL(ask.VerifiedPrice),
		)
	}

	return asks, nil
}

func GetAsks(ctx context.Context, api venusAPI.FullNode, capi clientapi.IMarketClient) ([]*storagemarket.StorageAsk, error) {
	miners, err := api.StateListMiners(ctx, vtypes.EmptyTSK)
	if err != nil {
		log.Errorf("getting miner list: %w", err)
		return nil, err
	}

	var withMinPower []address.Address

	for _, miner := range miners {
		power, err := api.StateMinerPower(ctx, miner, vtypes.EmptyTSK)
		if err != nil {
			continue
		}

		if power.HasMinPower {
			withMinPower = append(withMinPower, miner)
		}
	}

	var asks []*storagemarket.StorageAsk
	for _, miner := range withMinPower {
		mi, err := api.StateMinerInfo(ctx, miner, vtypes.EmptyTSK)
		if err != nil {
			continue
		}
		if mi.PeerId == nil {
			continue
		}

		ask, err := capi.ClientQueryAsk(ctx, *mi.PeerId, miner)
		if err != nil {
			continue
		}

		asks = append(asks, ask)
	}

	sort.Slice(asks, func(i, j int) bool {
		return asks[i].Price.LessThan(asks[j].Price)
	})
	return asks, nil
}
