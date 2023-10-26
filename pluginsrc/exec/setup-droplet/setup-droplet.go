package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/filecoin-project/go-address"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	droplet "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
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
	Droplet      droplet.DropletMarketDeployReturn `json:"Droplet" jsonschema:"Droplet" title:"Droplet" description:"droplet market return"`
	MinerAddress address.Address                   `json:"minerAddress" jsonschema:"minerAddress" title:"Miner Address" require:"true" description:"miner to set market address"`
}

type SetupDropletReturn struct {
	Multiaddrs         string
	PeerID             peer.ID
	SetAddrMessageId   string
	SetPeerIDMessageId string
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*SetupDropletReturn, error) {
	client, closer, err := marketapi.DialIMarketRPC(ctx, params.Droplet.SvcEndpoint.ToMultiAddr(), params.Droplet.UserToken, nil)
	if err != nil {
		log.Errorf("new market api failed: %v\n", err)
		return nil, err
	}
	defer closer()

	addrInfo, err := client.NetAddrsListen(ctx)
	fmt.Println("addr", addrInfo.Addrs, "peerid", addrInfo.ID)
	if err != nil || len(addrInfo.Addrs) == 0 {
		log.Errorf("client net addrs listen failed: %v\n", err)
		return nil, err
	}

	pods, err := droplet.GetPods(ctx, k8sEnv, params.Droplet.InstanceName)
	if err != nil {
		return nil, err
	}

	dns := "/dns4/" + params.Droplet.SVCName + "." + k8sEnv.NameSpace() + ".svc.cluster.local/tcp/58418"
	setAddrCmd := "./droplet actor set-addrs --miner=" + strings.TrimSpace(params.MinerAddress.String()) + " " + dns
	log.Infoln("setAddrCmd is: ", setAddrCmd)
	res, err := k8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), "/bin/sh", "-c", setAddrCmd)
	if err != nil {
		return nil, fmt.Errorf("exec ./droplet actor set-addrs failed")
	}

	log.Infoln("set-addr msg is: ", string(res))
	setAddrMessageId := string(res)[39:]
	log.Infoln("set-addr msg id: ", setAddrMessageId)
	if err != nil {
		return nil, err
	}

	setPeerIDCmd := "./droplet actor set-peer-id --miner=" + strings.TrimSpace(params.MinerAddress.String()) + " " + strings.TrimSpace(addrInfo.ID.String())
	log.Infoln("setPeerIDCmd is: ", setPeerIDCmd)
	res, err = k8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), "/bin/sh", "-c", setPeerIDCmd)
	if err != nil {
		return nil, fmt.Errorf("exec ./droplet actor set-peer-id failed")
	}

	log.Infoln("set-peer-id msg is: ", string(res))
	setPeerIDMessageId := string(res)[35:]
	log.Infoln("set-peer-id msg id: ", setPeerIDMessageId)
	if err != nil {
		return nil, err
	}

	return &SetupDropletReturn{
		Multiaddrs:         dns,
		PeerID:             addrInfo.ID,
		SetAddrMessageId:   setAddrMessageId,
		SetPeerIDMessageId: setPeerIDMessageId,
	}, nil
}
