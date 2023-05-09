package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-state-types/big"
	vtypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/filecoin-project/venus/venus-shared/types/market/client"
	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/go-address"
	clientapi "github.com/filecoin-project/venus/venus-shared/api/market/client"
	"github.com/filecoin-project/venus/venus-shared/api/wallet"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "storage-deals",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "storage-deals",
}

type TestCaseParams struct {
	fx.In
	K8sEnv                     *env.K8sEnvDeployer `json:"-"`
	VenusAuth                  env.IDeployer       `json:"-" svcname:"VenusAuth"`
	MarketClient               env.IDeployer       `json:"-" svcname:"MarketClient"`
	Venus                      env.IDeployer       `json:"-" svcname:"Venus"`
	VenusWallet                env.IDeployer       `json:"-" svcname:"VenusWallet"`
	VenusSectorManagerDeployer env.IDeployer       `json:"-" svcname:"VenusSectorManager"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	walletAddr, err := CreateWallet(ctx, params)
	if err != nil {
		fmt.Printf("create wallet failed: %v\n", err)
		return nil, err
	}

	minerAddr, err := CreateMiner(ctx, params, walletAddr)
	if err != nil {
		fmt.Printf("create miner failed: %v\n", err)
		return nil, err
	}

	minerInfo, err := GetMinerInfo(ctx, params, minerAddr)
	if err != nil {
		fmt.Printf("get miner info failed: %v\n", err)
		return nil, err
	}
	fmt.Println("miner info: %v", minerInfo)

	err = SubmitRetrievalRequest(ctx, params, minerAddr)
	if err != nil {
		fmt.Printf("storage asks query failed: %v\n", err)
		return nil, err
	}
	return env.NewSimpleExec(), nil
}

func CreateWallet(ctx context.Context, params TestCaseParams) (address.Address, error) {
	pods, err := params.VenusWallet.Pods(ctx)
	if err != nil {
		return address.Undef, err
	}

	svc, err := params.VenusWallet.Svc(ctx)
	if err != nil {
		return address.Undef, err
	}
	walletToken, err := env.ReadWalletToken(ctx, params.K8sEnv, pods[0].GetName())
	if err != nil {
		return address.Undef, fmt.Errorf("read wallet token failed: %w\n", err)
	}

	endpoint := params.VenusWallet.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return address.Undef, fmt.Errorf("port forward failed: %w\n", err)
		}
	}

	walletRpc, closer, err := wallet.DialIFullAPIRPC(ctx, endpoint.ToMultiAddr(), walletToken, nil)
	if err != nil {
		return address.Undef, fmt.Errorf("dial iFullAPI rpc failed: %w\n", err)
	}
	defer closer()

	password := "123456"
	err = walletRpc.SetPassword(ctx, password)
	if err != nil {
		return address.Undef, fmt.Errorf("set password failed: %w\n", err)
	}

	walletAddr, err := walletRpc.WalletNew(ctx, vtypes.KTBLS)
	if err != nil {
		return address.Undef, fmt.Errorf("create wallet failed: %w\n", err)
	}
	fmt.Printf("wallet: %v\n", walletAddr)

	return walletAddr, nil
}

func CreateMiner(ctx context.Context, params TestCaseParams, walletAddr address.Address) (address.Address, error) {
	cmd := []string{
		"./venus-sector-manager",
		"util",
		"miner",
		"create",
		"--sector-size=8MiB",
		"--exid=" + string(rune(rand.Intn(100000))),
	}
	cmd = append(cmd, "--from="+walletAddr.String())

	pods, err := params.VenusSectorManagerDeployer.Pods(ctx)
	if err != nil {
		return address.Undef, err
	}

	minerAddr, err := params.K8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), cmd...)
	if err != nil {
		return address.Undef, fmt.Errorf("exec remote cmd failed: %w\n", err)
	}

	addr, err := address.NewFromBytes(minerAddr)
	if err != nil {
		return address.Undef, err
	}
	return addr, nil
}

func GetMinerInfo(ctx context.Context, params TestCaseParams, minerAddr address.Address) (string, error) {
	getMinerCmd := []string{
		"./venus-sector-manager",
		"util",
		"miner",
		"info",
		minerAddr.String(),
	}

	pods, err := params.VenusSectorManagerDeployer.Pods(ctx)
	if err != nil {
		return "", err
	}

	minerInfo, err := params.K8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), getMinerCmd...)
	if err != nil {
		return "", fmt.Errorf("exec remote cmd failed: %w\n", err)
	}

	return string(minerInfo), nil
}

func SubmitRetrievalRequest(ctx context.Context, params TestCaseParams, minerAddr address.Address) error {
	endpoint := params.MarketClient.SvcEndpoint()
	if env.Debug {
		pods, err := params.MarketClient.Pods(ctx)
		if err != nil {
			return err
		}

		svc, err := params.MarketClient.Svc(ctx)
		if err != nil {
			return err
		}

		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return err
		}
	}
	api, closer, err := clientapi.NewIMarketClientRPC(ctx, endpoint.ToHttp(), nil)
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
