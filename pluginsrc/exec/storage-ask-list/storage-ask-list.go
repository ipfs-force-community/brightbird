package main

import (
	"context"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/venus/pkg/constants"
	chain "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	vtypes "github.com/filecoin-project/venus/venus-shared/types"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletmarket "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("storage-ask-list")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "storage-ask-list",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "droplet 查询 miner 挂单信息列表",
}

type TestCaseParams struct {
	Venus        venus.VenusDeployReturn                 `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	Auth         sophonauth.SophonAuthDeployReturn       `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Droplet      dropletmarket.DropletMarketDeployReturn `json:"Droplet" jsonschema:"Droplet" title:"Droplet" description:"droplet return"`
	MinerAddress address.Address                         `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true"`
}

type StorageAskListReturn []*storagemarket.StorageAsk

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (StorageAskListReturn, error) {
	client, closer, err := marketapi.DialIMarketRPC(ctx, params.Droplet.SvcEndpoint.ToMultiAddr(), params.Droplet.UserToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	fnapi, closer, err := chain.DialFullNodeRPC(ctx, params.Venus.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	head, err := fnapi.ChainHead(ctx)
	if err != nil {
		return nil, err
	}

	asks, err := client.MarketListStorageAsk(ctx)
	if err != nil {
		return nil, err
	}

	log.Debugf("Miner\tPrice per GiB/Epoch\tVerified\tMin. Piece Size (padded)\tMax. Piece Size (padded)\tExpiry (Epoch)\tExpiry (Appx. Rem. Time)\tSeq. No.\n")

	var storageAskList StorageAskListReturn
	for _, sask := range asks {
		if sask != nil && sask.Ask != nil {
			storageAskList = append(storageAskList, sask.Ask)
		}

		dlt := sask.Ask.Expiry - head.Height()
		rem := "<expired>"
		if dlt > 0 {
			rem = (time.Second * time.Duration(int64(dlt)*int64(constants.MainNetBlockDelaySecs))).String()
		}

		ask := sask.Ask
		log.Debugf("%s\t%s\t%s\t%s\t%s\t%d\t%s\t%d\n", ask.Miner, vtypes.FIL(ask.Price), vtypes.FIL(ask.VerifiedPrice), vtypes.SizeStr(vtypes.NewInt(uint64(ask.MinPieceSize))), vtypes.SizeStr(vtypes.NewInt(uint64(ask.MaxPieceSize))), ask.Expiry, rem, ask.SeqNo)
	}

	return storageAskList, err
}
