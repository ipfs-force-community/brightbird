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

var log = logging.Logger("actor-delete")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "actor-delete",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "delete miner",
}

type TestCaseParams struct {
	Droplet      droplet.DropletMarketDeployReturn `json:"Droplet" jsonschema:"Droplet" title:"Droplet" description:"droplet return"`
	MinerAddress address.Address                   `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true" `
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
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
	deleteCmd := fmt.Sprintf("./droplet actor delete %v", minerAddr)
	log.Infoln("deleteCmd is: ", deleteCmd)
	res, err := k8sEnv.ExecRemoteCmd(ctx, execPodName, "/bin/sh", "-c", deleteCmd)
	if err != nil {
		return err
	}
	log.Infoln("actor delete success: ", string(res))

	// print config file
	// grep -A 2 '\[Miners\]'  ~/.droplet/config.toml
	printConfigCmd := `res=$(grep -B 1 -A 5 '\[Miners\]' ~/.droplet/config.toml ) && echo "------begin------	[~/.droplet/config.toml]" && echo "$res" && echo "------end--------"`
	log.Infoln("printConfigCmd is: ", printConfigCmd)
	if res, err := k8sEnv.ExecRemoteCmd(ctx, execPodName, "/bin/sh", "-c", printConfigCmd); err != nil {
		log.Errorf("printConfigCmd failed: %v\n", err)
	} else {
		log.Infoln("printConfigCmd result: \n", string(res))
	}
	// if ! grep -q "$search_term" <<< "$string"; then
	// check in actor list
	address.CurrentNetwork = address.Mainnet
	mMain := params.MinerAddress.String()
	address.CurrentNetwork = address.Testnet
	mTest := params.MinerAddress.String()
	checkInListCmd := fmt.Sprintf(`./droplet actor list | if grep -Eq "%s|%s"; then (exit 1); fi`, mMain, mTest)
	log.Infoln("checkInListCmd is: ", checkInListCmd)
	if _, err := k8sEnv.ExecRemoteCmd(ctx, execPodName, "/bin/sh", "-c", checkInListCmd); err != nil {
		return fmt.Errorf("actor %s not in actor list : %v", minerAddr, err)
	}

	log.Infof("get actor %v in list", minerAddr)

	return nil

}

// func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
// 	client, closer, err := marketapi.DialIMarketRPC(ctx, params.Droplet.SvcEndpoint.ToMultiAddr(), params.Droplet.UserToken, nil)
// 	if err != nil {
// 		return err
// 	}
// 	defer closer()

// 	if err := client.ActorDelete(ctx, params.MinerAddress); err != nil {
// 		return err
// 	}

// 	if err := actorList(ctx, params, client); err != nil {
// 		log.Errorln("list actor err %w", err)
// 		return err
// 	}

// 	log.Debugf("delete miner %s success\n", params.MinerAddress)
// 	return nil
// }
// func actorList(ctx context.Context, params TestCaseParams, api marketapi.IMarket) error {
// 	miners, err := api.ActorList(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	for _, miner := range miners {
// 		if miner.Addr == params.MinerAddress {
// 			return err
// 		}
// 	}

// 	return nil
// }
