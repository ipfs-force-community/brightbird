package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"
	chain "github.com/filecoin-project/venus/venus-shared/api/chain/v1"

	vtypes "github.com/filecoin-project/venus/venus-shared/types"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "wait-power",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "user miner add",
}

type TestCaseParams struct {
	Venus   venus.VenusDeployReturn           `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	Auth    sophonauth.SophonAuthDeployReturn `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Miner   address.Address                   `json:"miner" jsonschema:"miner" title:"Miner Address" require:"true" description:"miner address"`
	Timeout string                            `json:"timeout" jsonschema:"timeout" title:"Timeout" default:"60m" require:"true" description:"time to wait power default to 1h"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	dur, err := time.ParseDuration(params.Timeout)
	if err != nil {
		return err
	}

	chainRPC, closer, err := chain.DialFullNodeRPC(ctx, params.Venus.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	timout := time.NewTimer(dur)
	defer timout.Stop()

	tm := time.NewTicker(time.Second * 10)
	defer tm.Stop()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("cancel by context")
		case <-timout.C:
			return fmt.Errorf("timeout for wait context")
		case <-tm.C:
			power, err := chainRPC.StateMinerPower(ctx, params.Miner, vtypes.EmptyTSK)
			if err != nil {
				return err
			}

			if power.MinerPower.RawBytePower.GreaterThan(big.Zero()) {
				return nil
			}
		}

	}
}
