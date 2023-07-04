package main

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/filecoin-project/go-address"
	v1api "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	clientapi "github.com/filecoin-project/venus/venus-shared/api/market/client"
	vtypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
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
	fx.In
	K8sEnv          *env.K8sEnvDeployer `json:"-"`
	SophonAuth      env.IDeployer       `json:"-" svcname:"SophonAuth"`
	MarketClient    env.IDeployer       `json:"-" svcname:"MarketClient"`
	Venus           env.IDeployer       `json:"-" svcname:"Venus"`
	DamoclesManager env.IDeployer       `json:"-" svcname:"DamoclesManager"`
	CreateMiner     env.IExec           `json:"-" svcname:"CreateMiner"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {

	minerAddr, err := params.CreateMiner.Param("Miner")
	if err != nil {
		return nil, err
	}

	addr, err := env.UnmarshalJSON[address.Address](minerAddr.Raw())
	if err != nil {
		panic(err)
	}

	minerInfo, err := GetMinerInfo(ctx, params, addr)
	if err != nil {
		fmt.Printf("get miner info failed: %v\n", err)
		return nil, err
	}
	log.Infof("miner info: %v", minerInfo)

	err = StorageAsksQuery(ctx, params, addr)
	if err != nil {
		fmt.Printf("storage asks query failed: %v\n", err)
		return nil, err
	}
	return env.NewSimpleExec(), nil
}

func GetMinerInfo(ctx context.Context, params TestCaseParams, minerAddr address.Address) (string, error) {
	getMinerCmd := []string{
		"./venus-sector-manager",
		"util",
		"miner",
		"info",
		minerAddr.String(),
	}

	pods, err := params.DamoclesManager.Pods(ctx)
	if err != nil {
		return "", err
	}

	minerInfo, err := params.K8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), getMinerCmd...)
	if err != nil {
		return "", fmt.Errorf("exec remote cmd failed: %w", err)
	}

	return string(minerInfo), nil
}

func StorageAsksQuery(ctx context.Context, params TestCaseParams, maddr address.Address) error {
	endpoint, err := params.MarketClient.SvcEndpoint()
	if err != nil {
		return err
	}
	api, closer, err := clientapi.NewIMarketClientRPC(ctx, endpoint.ToHTTP(), nil)
	if err != nil {
		return err
	}
	defer closer()

	venusEndpoint, err := params.Venus.SvcEndpoint()
	if err != nil {
		return err
	}

	fnapi, closer, err := v1api.NewFullNodeRPC(ctx, venusEndpoint.ToHTTP(), nil)
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
