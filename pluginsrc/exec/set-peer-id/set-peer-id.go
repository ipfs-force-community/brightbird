package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/filecoin-project/go-address"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"

	droplet "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
)

var log = logging.Logger(Info.Name)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "set-peer-id",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "set droplet peer-id",
}

type TestCaseParams struct {
	Droplet      droplet.DropletMarketDeployReturn `json:"Droplet" jsonschema:"Droplet" title:"Droplet" description:"droplet market return"`
	MinerAddress address.Address                   `json:"minerAddress" jsonschema:"minerAddress" title:"Miner Address" require:"true" description:"miner to set market address"`
}

type TestReturn struct {
	PeerID             peer.ID
	SetPeerIDMessageId string
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*TestReturn, error) {
	client, closer, err := marketapi.DialIMarketRPC(ctx, params.Droplet.SvcEndpoint.ToMultiAddr(), params.Droplet.UserToken, nil)
	if err != nil {
		log.Errorf("new market api failed: %v\n", err)
		return nil, err
	}
	defer closer()

	addrInfo, err := client.NetAddrsListen(ctx)
	if err != nil || len(addrInfo.Addrs) == 0 {
		log.Errorf("client net addrs listen failed: %v\n", err)
		return nil, err
	}
	log.Infof("addrInfo: %v\n", addrInfo)

	pods, err := droplet.GetPods(ctx, k8sEnv, params.Droplet.InstanceName)
	if err != nil {
		return nil, err
	}
	if len(pods) == 0 {
		return nil, fmt.Errorf("get droplet pods failed")
	}
	execPodName := pods[0].GetName()
	log.Infof("exec pod name: %s", execPodName)

	// ./droplet actor set-peer-id --miner=t01004 12D3KooWJ4v1x2z5Z6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6
	setPeerIDCmd := fmt.Sprintf(
		"./droplet actor set-peer-id --miner=%s %s",
		strings.TrimSpace(params.MinerAddress.String()),
		strings.TrimSpace(addrInfo.ID.String()),
	)
	log.Infoln("setPeerIDCmd is: ", setPeerIDCmd)
	res, err := k8sEnv.ExecRemoteCmd(ctx, execPodName, "/bin/sh", "-c", setPeerIDCmd)
	if err != nil {
		return nil, fmt.Errorf("exec ./droplet actor set-peer-id failed: %v", err)
	}
	log.Infoln("set-peer-id msg is: ", string(res))
	setPeerIDMessageId := string(res)[35:]
	log.Infoln("set-peer-id msg id: ", setPeerIDMessageId)

	return &TestReturn{
		PeerID:             addrInfo.ID,
		SetPeerIDMessageId: setPeerIDMessageId,
	}, nil
}
