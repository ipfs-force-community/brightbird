package main

import (
	"context"

	"github.com/filecoin-project/go-address"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	droplet "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("actor-upsert")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "actor-upsert",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "更新或者新增miner",
}

type TestCaseParams struct {
	Droplet      droplet.DropletMarketDeployReturn `json:"Droplet" jsonschema:"Droplet" title:"Droplet" description:"droplet return"`
	MinerAddress address.Address                   `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true"`
	Account      string                            `json:"account"  jsonschema:"account" title:"account" require:"false" description:"create username"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	// ./droplet actor upsert --account admin-test-1 t01004
	pods, err := droplet.GetPods(ctx, k8sEnv, params.Droplet.InstanceName)
	if err != nil {
		return err
	}
	upsertCmd := "./droplet actor upsert --account " + params.Account + " " + params.MinerAddress.String()
	log.Infoln("upsertCmd is: ", upsertCmd)

	res, err := k8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), "/bin/sh", "-c", upsertCmd)
	if err != nil {
		return err
	}
	log.Infoln("actor upsert success: ", string(res))
	return nil
}
