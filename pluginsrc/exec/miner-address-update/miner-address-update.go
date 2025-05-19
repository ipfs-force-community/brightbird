package main

import (
	"context"
	"encoding/json"

	miner "github.com/ipfs-force-community/sophon-miner/api/client"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"

	sophonminer "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-miner"
)

var log = logging.Logger("storage-ask")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "miner-address-update",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "更新miner地址",
}

type TestCaseParams struct {
	SophonMiner sophonminer.SophonMinerDeployReturn `json:"SophonMiner"  jsonschema:"SophonMiner" title:"Sophon Miner" description:"sophon miner eturn" require:"true"`
	Limit       int64                               `json:"Limit"  jsonschema:"Limit" title:"Limit" require:"false"`
	Skip        int64                               `json:"Skip"  jsonschema:"Skip" title:"Skip" require:"false"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	api, closer, err := miner.NewMinerRPC(ctx, params.SophonMiner.SvcEndpoint.ToMultiAddr(), nil)
	if err != nil {
		return err
	}
	defer closer()

	miners, err := api.UpdateAddress(ctx, params.Skip, params.Limit)
	if err != nil {
		return err
	}

	formatJSON, err := json.MarshalIndent(miners, "", "\t")
	if err != nil {
		return err
	}
	log.Debugln(string(formatJSON))

	return nil
}
