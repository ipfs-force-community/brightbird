package main

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	abiPower "github.com/filecoin-project/go-state-types/builtin/v12/power"

	"github.com/docker/go-units"
	"github.com/filecoin-project/go-address"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	sophonmessager "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-messager"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"

	logging "github.com/ipfs/go-log/v2"

	"github.com/filecoin-project/venus/venus-shared/actors"
	"github.com/filecoin-project/venus/venus-shared/actors/builtin/miner"
	"github.com/filecoin-project/venus/venus-shared/actors/builtin/power"
	chain "github.com/filecoin-project/venus/venus-shared/api/chain/v1"

	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
)

var log = logging.Logger("add-miner")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "create_miner",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "create miner address",
}

type TestCaseParams struct {
	Venus    venus.VenusDeployReturn             `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	Auth     sophonauth.SophonAuthDeployReturn   `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Messager sophonmessager.SophonMessagerReturn `json:"SophonMessager"  jsonschema:"SophonMessager"  title:"Sophon Messager" require:"true" description:"messager return"`
	//todo support set owner/worker/controller

	Size       string          `json:"size" jsonschema:"size" title:"Miner SIze" require:"2KiB" description:"miner size (2Kib 8Mib 512Mib 32Gib 64Gib)"`
	WalletAddr address.Address `json:"walletAddr" jsonschema:"walletAddr" title:"Wallet Address" require:"true" description:"owner/worker address must be f3 address"`
	Confidence int             `json:"confidence"  jsonschema:"confidence"  title:"Confidence" default:"5" require:"true" description:"confience height for wait message"`
}

type CreateMinerReturn struct {
	Miner  address.Address `json:"miner" jsonschema:"miner" title:"Miner Address" require:"true" description:"miner address"`
	Owner  address.Address `json:"owner" jsonschema:"owner" title:"Owner Address" require:"true" description:"owner address of miner"`
	Worker address.Address `json:"worker" jsonschema:"worker" title:"Worker Address" require:"true" description:"worker address of miner"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*CreateMinerReturn, error) {
	minerAddr, err := CreateMiner(ctx, k8sEnv, params, params.WalletAddr)
	if err != nil {
		fmt.Printf("create miner failed: %v\n", err)
		return nil, err
	}

	return &CreateMinerReturn{
		Miner:  minerAddr,
		Owner:  params.WalletAddr,
		Worker: params.WalletAddr,
	}, nil
}

func CreateMiner(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams, walletAddr address.Address) (address.Address, error) {
	chainRPC, closer, err := chain.DialFullNodeRPC(ctx, params.Venus.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		return address.Undef, err
	}
	defer closer()

	messagerRPC, closer, err := messager.DialIMessagerRPC(ctx, params.Messager.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		return address.Undef, err
	}
	defer closer()

	ts, err := chainRPC.ChainHead(ctx)
	if err != nil {
		return address.Undef, fmt.Errorf("get chain head: %w", err)
	}

	tsk := ts.Key()

	nv, err := chainRPC.StateNetworkVersion(ctx, tsk)
	if err != nil {
		return address.Undef, fmt.Errorf("get network version: %w", err)
	}

	ssize, err := units.RAMInBytes(params.Size)
	if err != nil {
		return address.Undef, fmt.Errorf("failed to parse sector size: %w", err)
	}

	sealProof, err := miner.SealProofTypeFromSectorSize(abi.SectorSize(ssize), nv)
	if err != nil {
		return address.Undef, fmt.Errorf("invalid sector size %d: %w", ssize, err)
	}

	fromStr := walletAddr.String()
	from, err := ShouldAddress(fromStr, true, false)
	if err != nil {
		return address.Undef, fmt.Errorf("parse from addr %s: %w", fromStr, err)
	}

	actor, err := chainRPC.StateLookupID(ctx, from, tsk)
	if err != nil {
		return address.Undef, fmt.Errorf("lookup actor address: %w", err)
	}

	mlog := log.With("size", params.Size, "from", fromStr, "actor", actor.String())
	mlog.Info("constructing message")

	owner := actor
	worker := owner

	var pid abi.PeerID
	var multiaddrs []abi.Multiaddrs
	postProof, err := sealProof.RegisteredWindowPoStProofByNetworkVersion(nv)
	if err != nil {
		return address.Undef, fmt.Errorf("invalid seal proof type %d: %w", sealProof, err)
	}

	serializeParams, err := actors.SerializeParams(&abiPower.CreateMinerParams{
		Owner:               owner,
		Worker:              worker,
		WindowPoStProofType: postProof,
		Peer:                pid,
		Multiaddrs:          multiaddrs,
	})

	if err != nil {
		return address.Undef, fmt.Errorf("serialize params: %w", err)
	}

	messagerId, err := messagerRPC.PushMessage(ctx, &vTypes.Message{
		To:     power.Address,
		From:   from,
		Method: power.Methods.CreateMiner,
		Params: serializeParams,
	}, nil)
	if err != nil {
		return address.Undef, err
	}

	result, err := messagerRPC.WaitMessage(ctx, messagerId, uint64(params.Confidence))
	if err != nil {
		return address.Undef, err
	}

	if result.Receipt.ExitCode != 0 {
		return address.Undef, fmt.Errorf("message fail %d", result.Receipt.ExitCode)
	}

	r := &abiPower.CreateMinerReturn{}
	err = r.UnmarshalCBOR(bytes.NewReader(result.Receipt.Return))
	if err != nil {
		return address.Undef, err
	}

	mlog.Infof("new miner created %s", r.IDAddress.String())
	return r.IDAddress, nil

}

var ErrEmptyAddressString = fmt.Errorf("empty address string")

func ShouldAddress(s string, checkEmpty bool, allowActor bool) (address.Address, error) {
	if checkEmpty && s == "" {
		return address.Undef, ErrEmptyAddressString
	}

	if allowActor {
		id, err := strconv.ParseUint(s, 10, 64)
		if err == nil {
			return address.NewIDAddress(id)
		}
	}

	return address.NewFromString(s)
}
