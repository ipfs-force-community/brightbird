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

var log = logging.Logger("storage-ask-get-miner")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "storage-ask-get-miner",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "droplet 查询 miner 挂单信息",
}

type TestCaseParams struct {
	Venus        venus.VenusDeployReturn                 `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	Auth         sophonauth.SophonAuthDeployReturn       `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Droplet      dropletmarket.DropletMarketDeployReturn `json:"Droplet" jsonschema:"Droplet" title:"Droplet" description:"droplet return"`
	MinerAddress address.Address                         `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true"`
}

type AskGetMinerReturn *storagemarket.StorageAsk

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (AskGetMinerReturn, error) {
	client, closer, err := marketapi.DialIMarketRPC(ctx, params.Droplet.SvcEndpoint.ToMultiAddr(), params.Droplet.UserToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	sask, err := client.MarketGetAsk(ctx, params.MinerAddress)
	if err != nil {
		return nil, err
	}

	var ask *storagemarket.StorageAsk
	if sask != nil && sask.Ask != nil {
		ask = sask.Ask
	}
	if ask == nil {
		log.Debugln("miner does not have an ask")
	}

	fnapi, closer, err := chain.DialFullNodeRPC(ctx, params.Venus.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	head, err := fnapi.ChainHead(ctx)
	if err != nil {
		return nil, err
	}

	dlt := ask.Expiry - head.Height()
	rem := "<expired>"
	if dlt > 0 {
		rem = (time.Second * time.Duration(int64(dlt)*int64(constants.MainNetBlockDelaySecs))).String()
	}

	log.Debugln("Price per GiB/Epoch: ", vtypes.FIL(ask.Price))
	log.Debugln("Verified: ", vtypes.FIL(ask.VerifiedPrice))
	log.Debugln("Min. Piece Size (padded): ", vtypes.SizeStr(vtypes.NewInt(uint64(ask.MinPieceSize))))
	log.Debugln("Max. Piece Size (padded): ", vtypes.SizeStr(vtypes.NewInt(uint64(ask.MaxPieceSize))))
	log.Debugln("Expiry (Epoch): ", ask.Expiry)
	log.Debugln("Expiry (Appx. Rem. Time): ", rem)
	log.Debugln("Seq. No.: ", ask.SeqNo)

	return ask, err
}
