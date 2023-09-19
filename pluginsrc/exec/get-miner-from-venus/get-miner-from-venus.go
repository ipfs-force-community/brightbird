package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"
	chain "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	vtypes "github.com/filecoin-project/venus/venus-shared/types"
	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multiaddr"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("get-miner-from-venus")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "get-miner-from-venus",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "从venus检查能否获取到期望的miner",
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

	log.Debugf("Available Balance: %s\n", vtypes.FIL(availableBalance))
	log.Debugf("Owner:\t%s\n", mi.Owner)
	log.Debugf("Worker:\t%s\n", mi.Worker)
	log.Debugf("PeerID:\t%s\n", mi.PeerId)
	log.Debugf("Multiaddrs:\t")
	for _, addr := range mi.Multiaddrs {
		a, err := multiaddr.NewMultiaddrBytes(addr)
		if err != nil {
			return nil, fmt.Errorf("undecodable listen address: %v", err)
		}
		log.Debugf("%s ", a)
	}

	log.Debugf("Consensus Fault End:\t%d\n", mi.ConsensusFaultElapsed)
	log.Debugf("SectorSize:\t%s (%d)\n", vtypes.SizeStr(big.NewInt(int64(mi.SectorSize))), mi.SectorSize)
	pow, err := chainRPC.StateMinerPower(ctx, params.MinerAddress, ts.Key())
	if err != nil {
		return nil, err
	}

	rpercI := big.Div(big.Mul(pow.MinerPower.RawBytePower, big.NewInt(1000000)), pow.TotalPower.RawBytePower)
	qpercI := big.Div(big.Mul(pow.MinerPower.QualityAdjPower, big.NewInt(1000000)), pow.TotalPower.QualityAdjPower)

	log.Debugf("Byte Power:   %s / %s (%0.4f%%)\n",
		vtypes.SizeStr(pow.MinerPower.RawBytePower),
		vtypes.SizeStr(pow.TotalPower.RawBytePower),
		float64(rpercI.Int64())/10000)

	log.Debugf("Actual Power: %s / %s (%0.4f%%)\n",
		vtypes.DeciStr(pow.MinerPower.QualityAdjPower),
		vtypes.DeciStr(pow.TotalPower.QualityAdjPower),
		float64(qpercI.Int64())/10000)

	return &GetMinerInfoReturn{
		MinerInfo: mi,
	}, nil
}
