package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"

	droplet "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
)

var log = logging.Logger("actor-upsert")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "actor-upsert",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "update or new miner",
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
	if len(pods) == 0 {
		return fmt.Errorf("no droplet pod found")
	}
	execPodName := pods[0].GetName()

	minerAddr := params.MinerAddress.String()

	// upsert actor
	upsertCmd := "./droplet actor upsert --account " + params.Account + " " + minerAddr
	log.Infoln("upsertCmd is: ", upsertCmd)
	res, err := k8sEnv.ExecRemoteCmd(ctx, execPodName, "/bin/sh", "-c", upsertCmd)
	if err != nil {
		return err
	}
	log.Infoln("actor upsert success: ", string(res))

	// print config file
	// grep -A 2 '\[Miners\]'  ~/.droplet/config.toml
	printConfigCmd := `res=$(grep -B 1 -A 5 '\[Miners\]' ~/.droplet/config.toml ) && echo "------begin------	[~/.droplet/config.toml]" && echo "$res" && echo "------end--------"`
	log.Infoln("printConfigCmd is: ", printConfigCmd)
	if res, err := k8sEnv.ExecRemoteCmd(ctx, execPodName, "/bin/sh", "-c", printConfigCmd); err != nil {
		log.Errorf("printConfigCmd failed: %v\n", err)
	} else {
		log.Infoln("printConfigCmd result: \n", string(res))
	}

	// check in actor list
	address.CurrentNetwork = address.Mainnet
	mMain := params.MinerAddress.String()
	address.CurrentNetwork = address.Testnet
	mTest := params.MinerAddress.String()
	checkInListCmd := fmt.Sprintf(`./droplet actor list | grep -Eq "%s|%s"`, mMain, mTest)
	log.Infoln("checkInListCmd is: ", checkInListCmd)
	if _, err := k8sEnv.ExecRemoteCmd(ctx, execPodName, "/bin/sh", "-c", checkInListCmd); err != nil {
		return fmt.Errorf("actor %s not in actor list : %v", minerAddr, err)
	}

	log.Infof("get actor %v in list", minerAddr)

	return nil
}
