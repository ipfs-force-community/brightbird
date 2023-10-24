package main

import (
	"context"
	"strings"

	"github.com/filecoin-project/go-address"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletclient "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-client"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("client-asks-query-miner")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "client-asks-query-miner",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "droplet-client 查询 miner 挂单信息",
}

type TestCaseParams struct {
	DropletClient dropletclient.DropletClientDeployReturn `json:"DropClient" jsonschema:"DropClient" title:"DropletClient" description:"droplet client return"`
	MinerAddress  address.Address                         `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true" description:"minerAddress"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) ([]byte, error) {
	asksQueryCmd := "./droplet-client storage asks query " + strings.TrimSpace(params.MinerAddress.String())
	log.Infoln("asksQueryCmd is: ", asksQueryCmd)

	pods, err := dropletclient.GetPods(ctx, k8sEnv, params.DropletClient.InstanceName)
	if err != nil {
		return nil, err
	}

	minerInfo, err := k8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), "/bin/sh", "-c", asksQueryCmd)
	if err != nil {
		return nil, err
	}
	log.Infoln("minerInfo is: ", string(minerInfo))

	return minerInfo, nil
}
