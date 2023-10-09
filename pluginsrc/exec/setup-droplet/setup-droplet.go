package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/builtin"
	"github.com/filecoin-project/venus/venus-shared/actors"
	chain "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	vtypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	droplet "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	sophonmessager "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-messager"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("setup-droplet")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "setup-droplet",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "设置droplet的addrs和peer-id信息",
}

type TestCaseParams struct {
	Venus    venus.VenusDeployReturn             `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	Auth     sophonauth.SophonAuthDeployReturn   `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Messager sophonmessager.SophonMessagerReturn `json:"SophonMessager"  jsonschema:"SophonMessager"  title:"Sophon Messager" require:"true" description:"messager return"`
	Droplet  droplet.DropletMarketDeployReturn   `json:"Droplet" jsonschema:"Droplet" title:"Droplet" description:"droplet market return"`

	MinerAddress address.Address `json:"minerAddress" jsonschema:"minerAddress" title:"Miner Address" require:"true" description:"miner to set market address"`
	GasLimt      int64           `json:"gasLimt" jsonschema:"gasLimt" title:"gasLimt" require:"false" description:"set gas limit"`
	Confidence   int             `json:"confidence"  jsonschema:"confidence"  title:"Confidence" default:"5" require:"true" description:"confience height for wait message"`
}

type SetupDropletReturn struct {
	Multiaddrs         []abi.Multiaddrs
	PeerID             peer.ID
	SetAddrMessageId   string
	SetPeerIdMessageId string
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*SetupDropletReturn, error) {
	client, closer, err := marketapi.DialIMarketRPC(ctx, params.Droplet.SvcEndpoint.ToMultiAddr(), params.Droplet.UserToken, nil)
	if err != nil {
		log.Errorf("new market api failed: %v\n", err)
		return nil, err
	}
	defer closer()

	fapi, closer, err := chain.DialFullNodeRPC(ctx, params.Venus.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		log.Errorf("new venus api failed: %v\n", err)
		return nil, err
	}
	defer closer()

	addrInfo, err := client.NetAddrsListen(ctx)
	fmt.Println("addr", addrInfo.Addrs, "peerid", addrInfo.ID)
	if err != nil || len(addrInfo.Addrs) == 0 {
		log.Errorf("client net addrs listen failed: %v\n", err)
		return nil, err
	}

	var addrs []abi.Multiaddrs
	for _, a := range addrInfo.Addrs {
		if !strings.HasPrefix(a.String(), "/ip4/127.0.0.1/") && !strings.HasPrefix(a.String(), "/ip6/") {
			addrs = append(addrs, a.Bytes())
		}
	}
	addrMessageParams, err := actors.SerializeParams(&vtypes.ChangeMultiaddrsParams{NewMultiaddrs: addrs})
	if err != nil {
		return nil, err
	}

	pid := addrInfo.ID
	pidMessageParams, err := actors.SerializeParams(&vtypes.ChangePeerIDParams{NewID: abi.PeerID(pid)})
	if err != nil {
		log.Errorf("Construct params peer-id failed: %v\n", err)
		return nil, err
	}

	setAddrMessageId, err := SendMessage(ctx, params, addrMessageParams, client, fapi, builtin.MethodsMiner.ChangeMultiaddrs)
	if err != nil || setAddrMessageId == "" {
		log.Errorf("set address failed: %v\n", err)
		return nil, err
	}

	setPeerIdMessageId, err := SendMessage(ctx, params, pidMessageParams, client, fapi, builtin.MethodsMiner.ChangePeerID)
	if err != nil || setPeerIdMessageId == "" {
		log.Errorln("set peer-id failed: %v\n", err)
		return nil, err
	}

	return &SetupDropletReturn{
		Multiaddrs:         addrs,
		PeerID:             pid,
		SetAddrMessageId:   setAddrMessageId,
		SetPeerIdMessageId: setPeerIdMessageId,
	}, nil
}

func SendMessage(ctx context.Context, params TestCaseParams, messageParams []byte, client marketapi.IMarket, fapi chain.FullNode, method abi.MethodNum) (string, error) {
	minfo, err := fapi.StateMinerInfo(ctx, params.MinerAddress, vtypes.EmptyTSK)
	if err != nil {
		return "", err
	}

	messageId, err := client.MessagerPushMessage(ctx, &vtypes.Message{
		To:       params.MinerAddress,
		From:     minfo.Worker,
		Value:    vtypes.NewInt(0),
		GasLimit: params.GasLimt,
		Method:   method,
		Params:   messageParams,
	}, nil)
	if err != nil {
		log.Errorf("push message failed: %v\n", err)
		return "", err
	}

	log.Debugf("Requested multiaddrs change in message %s\n", messageId)

	messagerRPC, closer, err := messager.DialIMessagerRPC(ctx, params.Messager.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		return "", err
	}
	defer closer()

	result, err := messagerRPC.WaitMessage(ctx, messageId.String(), uint64(params.Confidence))
	if err != nil {
		return "", err
	}

	if result.Receipt.ExitCode != 0 {
		log.Errorln("message fail %d", result.Receipt.ExitCode)
		return "", err
	}

	return messageId.String(), nil
}
