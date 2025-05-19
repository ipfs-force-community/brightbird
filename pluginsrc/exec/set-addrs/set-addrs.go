package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/filecoin-project/go-address"
	logging "github.com/ipfs/go-log/v2"

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
	Name:        "set-addrs",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "set droplet addrs ",
}

type TestCaseParams struct {
	Droplet      droplet.DropletMarketDeployReturn `json:"Droplet" jsonschema:"Droplet" title:"Droplet" description:"droplet market return"`
	MinerAddress address.Address                   `json:"minerAddress" jsonschema:"minerAddress" title:"Miner Address" require:"true" description:"miner to set market address"`
}

type TestReturn struct {
	Multiaddrs       string
	SetAddrMessageId string
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*TestReturn, error) {
	pods, err := droplet.GetPods(ctx, k8sEnv, params.Droplet.InstanceName)
	if err != nil {
		return nil, err
	}
	if len(pods) == 0 {
		return nil, fmt.Errorf("get droplet pods failed")
	}
	execPodName := pods[0].GetName()
	log.Infof("exec pod name: %s", execPodName)

	// ./droplet actor set-addrs --miner=t01004 /dns4/droplet-market-0.droplet-market.svc.cluster.local/tcp/58418
	dns := fmt.Sprintf("/dns4/%s.%s.svc.cluster.local/tcp/58418", params.Droplet.SVCName, k8sEnv.NameSpace())
	setAddrCmd := fmt.Sprintf(
		"./droplet actor set-addrs --miner=%s  %s",
		strings.TrimSpace(params.MinerAddress.String()),
		dns,
	)
	log.Infoln("setAddrCmd is: ", setAddrCmd)
	res, err := k8sEnv.ExecRemoteCmd(ctx, execPodName, "/bin/sh", "-c", setAddrCmd)
	if err != nil {
		return nil, fmt.Errorf("exec [./droplet actor set-addrs] failed: %v ", err)
	}
	log.Infoln("set-addrs msg is: ", string(res))
	setAddrMessageId := string(res)[39:]
	log.Infoln("set-addrs msg id: ", setAddrMessageId)

	return &TestReturn{
		Multiaddrs:       dns,
		SetAddrMessageId: setAddrMessageId,
	}, nil
}
