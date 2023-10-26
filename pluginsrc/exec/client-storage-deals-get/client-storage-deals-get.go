package main

import (
	"context"
	"fmt"
	"time"

	"github.com/filecoin-project/go-fil-markets/storagemarket"
	clientapi "github.com/filecoin-project/venus/venus-shared/api/market/client"
	"github.com/filecoin-project/venus/venus-shared/types/market/client"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletclient "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-client"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("client-storage-deals-get")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "client-storage-deals-get",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "在droplet-client检索订单的状态是否符合预期",
}

type TestCaseParams struct {
	DropletClient dropletclient.DropletClientDeployReturn `json:"DropletClient" jsonschema:"DropletClient" title:"DropletClient" require:"true" description:"droplet client return"`

	State       storagemarket.StorageDealStatus `json:"state" jsonschema:"state" title:"state" require:"true" description:"13-CheckForAcceptance, 29-AwaitingPreCommit"`
	ProposalCid string                          `json:"ProposalCid" jsonschema:"ProposalCid" title:"ProposalCid" require:"true" description:"ProposalCid"`
	TimeOut     int                             `json:"TimeOut" jsonschema:"TimeOut" title:"TimeOut" require:"true" default:"10000" description:"TimeOut"`
}

type ClientStorageDealsList = []client.DealInfo

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*ClientStorageDealsList, error) {
	api, closer, err := clientapi.DialIMarketClientRPC(ctx, params.DropletClient.SvcEndpoint.ToMultiAddr(), params.DropletClient.ClientToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	var finalDeals []client.DealInfo

	startTime := time.Now()
	timeout := time.Duration(params.TimeOut) * time.Second

	for {
		localDeals, err := api.ClientListDeals(ctx)
		if err != nil {
			return nil, err
		}
		log.Debugln("params.DealCid: ", params.ProposalCid)
		for _, deal := range localDeals {
			log.Debugln("deal.ProposalCid: ", deal.ProposalCid)
			if params.ProposalCid == deal.ProposalCid.String() {
				finalDeals = append(finalDeals, deal)
			}
		}

		if len(finalDeals) > 0 {
			for _, deal := range finalDeals {
				log.Debugln("params.State: ", params.State)
				log.Debugln("deal.State: ", deal.State)
				if deal.State == storagemarket.StorageDealFailing || deal.State == storagemarket.StorageDealError {
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
		finalDeals = []client.DealInfo{}
	}
}
