package main

import (
	"context"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	damoclesmanager "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/damocles-manager"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "damocles-manager-add-miner",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "add a new miner to daocles manager",
}

type TestCaseParams struct {
	DamoclesManager damoclesmanager.DamoclesManagerReturn `json:"damoclesManager" jsonschema:"damoclesManager" title:"DamoclesManager" description:"damocles manager return"`
	damoclesmanager.MinerCfg
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	err := damoclesmanager.AddMiner(ctx, k8sEnv, params.DamoclesManager, params.MinerCfg)
	if err != nil {
		return err
	}

	client, closer, err := damoclesmanager.BuildDamoclesClient(ctx, params.DamoclesManager)
	if err != nil {
		return err
	}
	defer closer()

	_, err = client.GetMinerConfig(ctx, params.MinerActor)
	return err
}
