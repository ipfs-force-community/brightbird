package main

import (
	"context"
	"fmt"
	"time"

	"github.com/filecoin-project/go-address"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	"github.com/filecoin-project/venus/venus-shared/types/market"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletmarket "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("storage-deal-get")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "storage-deal-get",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "在droplet检索订单的状态是否符合预期",
}

type TestCaseParams struct {
	Droplet dropletmarket.DropletMarketDeployReturn `json:"Droplet" jsonschema:"Droplet" title:"Droplet" description:"droplet return"`

	State       uint64 `json:"state" jsonschema:"state" title:"state" require:"true" description:"18-StorageDealWaitingForData, 24-StorageDealPublish"`
	ProposalCid string `json:"ProposalCid" jsonschema:"ProposalCid" title:"ProposalCid" require:"true" description:"ProposalCid"`
	TimeOut     int    `json:"TimeOut" jsonschema:"TimeOut" title:"TimeOut" require:"true" default:"10" description:"TimeOut"`
}

type StorageDealsGetReturn = []market.MinerDeal

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*StorageDealsGetReturn, error) {
	client, closer, err := marketapi.DialIMarketRPC(ctx, params.Droplet.SvcEndpoint.ToMultiAddr(), params.Droplet.UserToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	listParams := market.StorageDealQueryParams{
		Miner:             address.Undef,
		State:             &params.State,
		Client:            "",
		DiscardFailedDeal: false,
		Page: market.Page{
			Offset: 0,
			Limit:  20,
		},
	}

	var finalDeals []market.MinerDeal

	startTime := time.Now()
	timeout := time.Duration(params.TimeOut) * time.Second

	for {
		deals, err := client.MarketListIncompleteDeals(ctx, &listParams)
		if err != nil {
			return nil, err
		}
		log.Debugln("params.DealCid: ", params.ProposalCid)
		for _, deal := range deals {
			log.Debugln("deal.ProposalCid: ", deal.ProposalCid)
			if params.ProposalCid == deal.ProposalCid.String() {
				finalDeals = append(finalDeals, deal)
			}
		}

		if len(finalDeals) > 0 {
			for _, deal := range finalDeals {
				log.Debugln("params.State: ", params.State)
				log.Debugln("deal.State: ", deal.State)
				// 26-StorageDealError, 11-StorageDealFailing
				if deal.State == 11 || deal.State == 26 {
					return nil, fmt.Errorf("deal failing or has error")
				} else if params.State == deal.State {
					return &finalDeals, nil
				}
			}
		}

		elapsedTime := time.Since(startTime)

		if elapsedTime >= timeout {
			log.Errorln("time out error")
			return nil, fmt.Errorf("time out error")
		}

		time.Sleep(1 * time.Second)
		finalDeals = []market.MinerDeal{}
	}
}
