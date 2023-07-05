package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	v1api "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	venusAPI "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	clientapi "github.com/filecoin-project/venus/venus-shared/api/market/client"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	vtypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	damoclesmanager "github.com/hunjixin/brightbird/pluginsrc/deploy/damocles-manager"
	dropletmarket "github.com/hunjixin/brightbird/pluginsrc/deploy/droplet-market"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/hunjixin/brightbird/pluginsrc/deploy/venus"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/peer"
)

var log = logging.Logger("storage-deals")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "storage-deals",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "storage-deals",
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

	err = StorageAsksQuery(ctx, params, params.MinerAddress)
	if err != nil {
		fmt.Printf("storage asks query failed: %v\n", err)
		return err
	}
	return nil
}

func StorageAsksQuery(ctx context.Context, params TestCaseParams, maddr address.Address) error {
	api, closer, err := clientapi.NewIMarketClientRPC(ctx, params.DropletMarket.SvcEndpoint.ToMultiAddr(), nil)
	if err != nil {
		return err
	}
	defer closer()

	fnapi, closer, err := v1api.NewFullNodeRPC(ctx, params.Venus.SvcEndpoint.ToMultiAddr(), nil)
	if err != nil {
		return err
	}
	defer closer()

	var pid peer.ID

	mi, err := fnapi.StateMinerInfo(ctx, maddr, vtypes.EmptyTSK)
	if err != nil {
		return fmt.Errorf("failed to get peerID for miner: %w", err)
	}

	if mi.PeerId == nil || *mi.PeerId == peer.ID("SETME") {
		return fmt.Errorf("the miner hasn't initialized yet")
	}

	pid = *mi.PeerId

	ask, err := api.ClientQueryAsk(ctx, pid, maddr)
	if err != nil {
		return err
	}

	fmt.Printf("Ask: %s\n", maddr)
	fmt.Printf("Price per GiB: %s\n", vtypes.FIL(ask.Price))
	fmt.Printf("Verified Price per GiB: %s\n", vtypes.FIL(ask.VerifiedPrice))
	fmt.Printf("Max Piece size: %s\n", vtypes.SizeStr(vtypes.NewInt(uint64(ask.MaxPieceSize))))
	fmt.Printf("Min Piece size: %s\n", vtypes.SizeStr(vtypes.NewInt(uint64(ask.MinPieceSize))))

	return nil
}
