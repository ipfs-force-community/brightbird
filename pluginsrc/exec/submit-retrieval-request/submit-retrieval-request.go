package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-state-types/big"
	clientapi "github.com/filecoin-project/venus/venus-shared/api/market/client"
	vtypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/filecoin-project/venus/venus-shared/types/market/client"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	damoclesmanager "github.com/hunjixin/brightbird/pluginsrc/deploy/damocles-manager"
	dropletmarket "github.com/hunjixin/brightbird/pluginsrc/deploy/droplet-market"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/hunjixin/brightbird/pluginsrc/deploy/venus"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"

	venusAPI "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
)

var log = logging.Logger("submit-retrieval-request")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "submit-retrieval-request",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "submit-retrieval-request",
}

type TestCaseParams struct {
	Auth            sophonauth.SophonAuthDeployReturn       `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Venus           venus.VenusDeployReturn                 `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	DropletMarket   dropletmarket.DropletMarketDeployReturn `json:"DropletMarket" jsonschema:"DropletMarket" title:"DropletMarket" description:"droplet market return"`
	DamoclesManager damoclesmanager.DamoclesManagerReturn   `json:"DamoclesManager" jsonschema:"DamoclesManager" title:"Damocles Manager" description:"damocles manager return" require:"true"`
	MinerAddress    address.Address                         `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true" `
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	fullNode, closer, err := venusAPI.DialFullNodeRPC(ctx, params.Venus.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
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

	err = SubmitRetrievalRequest(ctx, params, params.MinerAddress)
	if err != nil {
		fmt.Printf("storage asks query failed: %v\n", err)
		return err
	}
	return nil
}

func SubmitRetrievalRequest(ctx context.Context, params TestCaseParams, minerAddr address.Address) error {
	api, closer, err := clientapi.NewIMarketClientRPC(ctx, params.DropletMarket.SvcEndpoint.ToMultiAddr(), nil)
	if err != nil {
		return err
	}
	defer closer()

	eref, err := retrieve(ctx, api, minerAddr)
	if err != nil {
		return err
	}

	path := "test.txt"
	err = api.ClientExport(ctx, *eref, client.FileRef{
		Path:  path,
		IsCAR: false,
	})

	if err != nil {
		return err
	}
	fmt.Println("Success")
	return nil
}

func retrieve(ctx context.Context, fapi clientapi.IMarketClient, minerAddr address.Address) (*client.ExportRef, error) {
	var payer address.Address
	var err error

	payer, err = fapi.DefaultAddress(ctx)

	if err != nil {
		return nil, err
	}

	fileName := "test.txt"
	file, err := cid.Parse(fileName)
	if err != nil {
		return nil, err
	}

	var pieceCid *cid.Cid

	var eref *client.ExportRef

	var offer client.QueryOffer

	offer, err = fapi.ClientMinerQueryOffer(ctx, minerAddr, file, pieceCid)
	if err != nil {
		return nil, err
	}

	if offer.Err != "" {
		return nil, fmt.Errorf("offer error: %s", offer.Err)
	}

	maxPrice := vtypes.MustParseFIL("0")

	if offer.MinPrice.GreaterThan(big.Int(maxPrice)) {
		return nil, fmt.Errorf("failed to find offer satisfying maxPrice: %s", maxPrice)
	}

	o := offer.Order(payer)
	var sel *client.DataSelector
	o.DataSelector = sel

	subscribeEvents, err := fapi.ClientGetRetrievalUpdates(ctx)
	if err != nil {
		return nil, fmt.Errorf("error setting up retrieval updates: %w", err)
	}
	retrievalRes, err := fapi.ClientRetrieve(ctx, o)
	if err != nil {
		return nil, fmt.Errorf("error setting up retrieval: %w", err)
	}

	start := time.Now()
readEvents:
	for {
		var evt client.RetrievalInfo
		select {
		case <-ctx.Done():
			return nil, errors.New("retrieval timed out")
		case evt = <-subscribeEvents:
			if evt.ID != retrievalRes.DealID {
				continue
			}
		}

		event := "New"
		if evt.Event != nil {
			event = retrievalmarket.ClientEvents[*evt.Event]
		}

		fmt.Printf("Recv %s, Paid %s, %s (%s), %s [%d|%d]\n",
			vtypes.SizeStr(vtypes.NewInt(evt.BytesReceived)),
			vtypes.FIL(evt.TotalPaid),
			strings.TrimPrefix(event, "ClientEvent"),
			strings.TrimPrefix(retrievalmarket.DealStatuses[evt.Status], "DealStatus"),
			time.Since(start).Truncate(time.Millisecond),
			evt.ID,
			vtypes.NewInt(evt.BytesReceived),
		)

		switch evt.Status {
		case retrievalmarket.DealStatusCompleted:
			break readEvents
		case retrievalmarket.DealStatusRejected:
			return nil, fmt.Errorf("retrieval Proposal Rejected: %s", evt.Message)
		case retrievalmarket.DealStatusCancelled,
			retrievalmarket.DealStatusDealNotFound,
			retrievalmarket.DealStatusErrored:
			return nil, fmt.Errorf("retrieval Error: %s", evt.Message)
		}
	}

	eref = &client.ExportRef{
		Root:   file,
		DealID: retrievalRes.DealID,
	}

	return eref, nil
}
