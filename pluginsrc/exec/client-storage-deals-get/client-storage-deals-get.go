package main

import (
	"context"
	"strconv"
	"time"

	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
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
	Description: "验证droplet在被检索时功能是否正常",
}

type TestCaseParams struct {
	DropletClient dropletclient.DropletClientDeployReturn `json:"DropletClient" jsonschema:"DropletClient" title:"DropletClient" require:"true" description:"droplet client return"`

	State   storagemarket.StorageDealStatus `json:"state" jsonschema:"state" title:"state" require:"true" description:"29-AwaitingPreCommit"`
	DealCid string                          `json:"DealCid" jsonschema:"DealCid" title:"DealCid" require:"true" description:"DealCid"`
	TimeOut int                             `json:"TimeOut" jsonschema:"TimeOut" title:"TimeOut" require:"true" default:"10" description:"TimeOut"`
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
	paramDealID := abi.DealID(DealCid)

	startTime := time.Now()
	timeout := time.Duration(params.TimeOut) * time.Second

	for {
		localDeals, err := api.ClientListOfflineDeals(ctx)
		if err != nil {
			return nil, err
		}

		for _, deal := range localDeals {
			if params.State == deal.State && paramDealID == deal.DealID {
				finalDeals = append(finalDeals, deal)
			}
		}
		if len(finalDeals) > 0 {
			break
		}

		elapsedTime := time.Since(startTime)

		if elapsedTime >= timeout {
			log.Errorln("time out error")
			break
		}

		time.Sleep(1 * time.Second)
		finalDeals = []client.DealInfo{}
	}

	return &finalDeals, nil
}
