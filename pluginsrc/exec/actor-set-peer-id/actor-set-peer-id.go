package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/builtin"
	"github.com/filecoin-project/venus/venus-shared/actors"
	v1api "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	vtypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	dropletmarket "github.com/hunjixin/brightbird/pluginsrc/deploy/droplet-market"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	sophongateway "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-gateway"
	sophonmessager "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-messager"
	"github.com/hunjixin/brightbird/pluginsrc/deploy/venus"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "actor-set-peer-id",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "actor set peer-id",
}

type TestCaseParams struct {
	SophonAuth     sophonauth.SophonAuthDeployReturn       `json:"SophonAuth"`
	Gateway        sophongateway.SophonGatewayReturn       `json:"Gateway"`
	Venus          venus.VenusDeployReturn                 `json:"Venus"`
	SophonMessager sophonmessager.SophonMessagerReturn     `json:"SophonMessager"`
	DropletMarket  dropletmarket.DropletMarketDeployReturn `json:"DropletMarket" description:"droplet market return "`
	MinerAddress   address.Address                         `json:"minerAddress" type:"string"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	messageId, err := SetActorAddr(ctx, params)
	if err != nil {
		return fmt.Errorf("set actor address failed: %w", err)
	}

	err = VertifyMessageIfVaild(ctx, params, messageId)
	if err != nil {
		return err
	}

	return nil
}

func VertifyMessageIfVaild(ctx context.Context, params TestCaseParams, messageId cid.Cid) error {
	client, closer, err := messager.DialIMessagerRPC(ctx, params.SophonMessager.SvcEndpoint.ToMultiAddr(), params.SophonAuth.AdminToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	msg, err := client.GetMessageBySignedCid(ctx, messageId)
	if err != nil {
		return err
	}
	fmt.Printf("Message: %v\n", msg)

	return nil
}

func SetActorAddr(ctx context.Context, params TestCaseParams) (cid.Cid, error) {
	client, closer, err := marketapi.NewIMarketRPC(ctx, params.DropletMarket.SvcEndpoint.ToMultiAddr(), nil)
	if err != nil {
		return cid.Undef, err
	}
	defer closer()

	addrs, err := client.NetAddrsListen(ctx)
	if err != nil && addrs.Addrs != nil {
		return cid.Undef, nil
	}

	pid := addrs.ID

	MessageParams, err := ConstructParams(pid)
	if err != nil {
		return cid.Undef, err
	}
	minfo, err := GetMinerInfo(ctx, params, params.MinerAddress)
	if err != nil {
		return cid.Undef, err
	}

	mid, err := client.MessagerPushMessage(ctx, &vtypes.Message{
		To:       params.MinerAddress,
		From:     minfo.Worker,
		Value:    vtypes.NewInt(0),
		GasLimit: 0,
		Method:   builtin.MethodsMiner.ChangeMultiaddrs,
		Params:   MessageParams,
	}, nil)
	if err != nil {
		return cid.Undef, err
	}

	fmt.Printf("Requested multiaddrs change in message %s\n", mid)

	return cid.Undef, err
}

func ConstructParams(pid peer.ID) (param []byte, err error) {

	params, err := actors.SerializeParams(&vtypes.ChangePeerIDParams{NewID: abi.PeerID(pid)})
	if err != nil {
		return nil, err
	}
	return params, nil
}

func GetMinerInfo(ctx context.Context, params TestCaseParams, maddr address.Address) (vtypes.MinerInfo, error) {
	client, closer, err := v1api.DialFullNodeRPC(ctx, params.Venus.SvcEndpoint.ToMultiAddr(), params.SophonAuth.AdminToken, nil)
	if err != nil {
		return vtypes.MinerInfo{}, err
	}
	defer closer()

	minfo, err := client.StateMinerInfo(ctx, maddr, vtypes.EmptyTSK)
	if err != nil {
		return vtypes.MinerInfo{}, err
	}
	return minfo, nil
}
