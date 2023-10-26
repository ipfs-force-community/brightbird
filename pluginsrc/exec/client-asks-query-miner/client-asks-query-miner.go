package main

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	venusAPI "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	clientapi "github.com/filecoin-project/venus/venus-shared/api/market/client"
	vtypes "github.com/filecoin-project/venus/venus-shared/types"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletclient "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-client"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("client-asks-query-miner")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "client-asks-query-miner",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "droplet-client 查询 miner 挂单信息",
}

type TestCaseParams struct {
	Venus         venus.VenusDeployReturn                 `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	DropletClient dropletclient.DropletClientDeployReturn `json:"DropClient" jsonschema:"DropClient" title:"DropletClient" description:"droplet client return"`
	MinerAddress  address.Address                         `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true" description:"minerAddress"`
}

type StorageAskReturn *storagemarket.StorageAsk

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (StorageAskReturn, error) {
	api, closer, err := clientapi.DialIMarketClientRPC(ctx, params.DropletClient.SvcEndpoint.ToMultiAddr(), params.DropletClient.ClientToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	fnapi, closer, err := venusAPI.DialFullNodeRPC(ctx, params.Venus.SvcEndpoint.ToMultiAddr(), params.Venus.AdminToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	mi, err := fnapi.StateMinerInfo(ctx, params.MinerAddress, vtypes.EmptyTSK)
	if err != nil {
		log.Errorln("failed to get peerID for miner: %w", err)
		return nil, err
	}

	if mi.PeerId == nil || *mi.PeerId == peer.ID("SETME") {
		log.Errorln("the miner hasn't initialized yet")
		return nil, err
	}

	pid := *mi.PeerId
	ask, err := api.ClientQueryAsk(ctx, pid, params.MinerAddress)
	if err != nil {
		log.Errorln("storage asks query failed: %v\n", err)
		return nil, err
	}

	log.Debug("Ask: %s\n", ask.Miner)
	log.Debug("Price per GiB: %s\n", vtypes.FIL(ask.Price))
	log.Debug("Verified Price per GiB: %s\n", vtypes.FIL(ask.VerifiedPrice))
	log.Debug("Max Piece size: %s\n", vtypes.SizeStr(vtypes.NewInt(uint64(ask.MaxPieceSize))))
	log.Debug("Min Piece size: %s\n", vtypes.SizeStr(vtypes.NewInt(uint64(ask.MinPieceSize))))

	return ask, nil
}
