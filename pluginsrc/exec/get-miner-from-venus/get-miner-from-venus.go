package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"
	chain "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	vtypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/multiformats/go-multiaddr"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"

	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "get-miner-from-venus",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "get miner status from venus",
}

type TestCaseParams struct {
	Auth         sophonauth.SophonAuthDeployReturn `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Venus        venus.VenusDeployReturn           `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	MinerAddress address.Address                   `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true" `
}

type GetMinerInfoReturn struct {
	MinerInfo vtypes.MinerInfo `json:"minerInfo" jsonschema:"minerInfo" title:"minerInfo" require:"true" description:"miner info"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*GetMinerInfoReturn, error) {
	chainRPC, closer, err := chain.DialFullNodeRPC(ctx, params.Venus.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	mi, err := chainRPC.StateMinerInfo(ctx, params.MinerAddress, vtypes.EmptyTSK)
	if err != nil {
		return nil, err
	}

	ts, err := chainRPC.ChainHead(ctx)
	if err != nil {
		return nil, err
	}

	availableBalance, err := chainRPC.StateMinerAvailableBalance(ctx, params.MinerAddress, ts.Key())
	if err != nil {
		return nil, fmt.Errorf("getting miner available balance: %w", err)
	}

	fmt.Printf("Available Balance: %s\n", vtypes.FIL(availableBalance))
	fmt.Printf("Owner:\t%s\n", mi.Owner)
	fmt.Printf("Worker:\t%s\n", mi.Worker)
	fmt.Printf("PeerID:\t%s\n", mi.PeerId)
	if len(mi.Multiaddrs) > 0 {
		fmt.Println("Multiaddrs:")
		for _, addr := range mi.Multiaddrs {
			a, err := multiaddr.NewMultiaddrBytes(addr)
			if err != nil {
				return nil, fmt.Errorf("undecodable listen address: %v", err)
			}
			fmt.Println(a)
		}
	}
	fmt.Printf("Consensus Fault End:\t%d\n", mi.ConsensusFaultElapsed)
	fmt.Printf("SectorSize:\t%s (%d)\n", vtypes.SizeStr(big.NewInt(int64(mi.SectorSize))), mi.SectorSize)

	pow, err := chainRPC.StateMinerPower(ctx, params.MinerAddress, ts.Key())
	if err != nil {
		return nil, err
	}

	rpercI := big.Div(big.Mul(pow.MinerPower.RawBytePower, big.NewInt(1000000)), pow.TotalPower.RawBytePower)
	qpercI := big.Div(big.Mul(pow.MinerPower.QualityAdjPower, big.NewInt(1000000)), pow.TotalPower.QualityAdjPower)

	fmt.Printf("Byte Power:   %s / %s (%0.4f%%)\n",
		vtypes.SizeStr(pow.MinerPower.RawBytePower),
		vtypes.SizeStr(pow.TotalPower.RawBytePower),
		float64(rpercI.Int64())/10000)

	fmt.Printf("Actual Power: %s / %s (%0.4f%%)\n",
		vtypes.DeciStr(pow.MinerPower.QualityAdjPower),
		vtypes.DeciStr(pow.TotalPower.QualityAdjPower),
		float64(qpercI.Int64())/10000)

	return &GetMinerInfoReturn{
		MinerInfo: mi,
	}, nil
}
