package main

import (
	"context"
	"strconv"
	"time"

	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	clientapi "github.com/filecoin-project/venus/venus-shared/api/market/client"
	"github.com/filecoin-project/venus/venus-shared/types/market/client"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletclient "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-client"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "client-storage-deals-list",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "验证droplet在被检索时功能是否正常",
}

type TestCaseParams struct {
	DropletClient dropletclient.DropletClientDeployReturn `json:"DropletClient" jsonschema:"DropletClient" title:"DropletClient" require:"true" description:"droplet client return"`

	State   storagemarket.StorageDealStatus `json:"state" jsonschema:"state" title:"state" require:"false" description:"state"`
	DealCid string                          `json:"DealCid" jsonschema:"DealCid" title:"DealCid" require:"false" description:"DealCid"`
}

type ClientStorageDealsList = []client.DealInfo

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*ClientStorageDealsList, error) {
	api, closer, err := clientapi.DialIMarketClientRPC(ctx, params.DropletClient.SvcEndpoint.ToMultiAddr(), params.DropletClient.ClientToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	var finalDeals []client.DealInfo

	DealCid, _ := strconv.ParseInt(params.DealCid, 10, 64)
	DealID := abi.DealID(DealCid)

	if params.State != storagemarket.StorageDealUnknown {
		for {
			localDeals, err := api.ClientListOfflineDeals(ctx)
			if err != nil {
				return nil, err
			}

			for _, deal := range localDeals {
				if deal.State == params.State {
					finalDeals = append(finalDeals, deal)
				}
			}

			if len(finalDeals) > 0 {
				break
			}

			time.Sleep(500 * time.Millisecond)
		}
	} else {
		localDeals, err := api.ClientListOfflineDeals(ctx)
		if err != nil {
			return nil, err
		}

		finalDeals = filterDeals(localDeals, params, DealID)
	}

	return &finalDeals, nil
}

func filterDeals(deals []client.DealInfo, params TestCaseParams, DealID abi.DealID) []client.DealInfo {
	var filteredDeals []client.DealInfo

	for _, deal := range deals {
		if params.State != storagemarket.StorageDealUnknown && deal.State != params.State {
			continue
		}
		if params.DealCid != "" && deal.DealID != DealID {
			continue
		}
		filteredDeals = append(filteredDeals, deal)
	}
	return filteredDeals
}
