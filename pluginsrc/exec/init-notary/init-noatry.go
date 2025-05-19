package main

import (
	"context"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"

	genesisnode "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/genesis-node"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "init-notary",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "init notary for genesis node",
}

type InputCaseParams struct {
	GenesisNodeRet genesisnode.GenesisReturn `json:"genesisNodeRet" jsonschema:"genesisNodeRet" title:"Genesis Node return Params" require:"true" description:"genesis node return params, used to get genesis node info"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params InputCaseParams) error {
	return genesisnode.InitNotary(ctx, k8sEnv, params.GenesisNodeRet.InstanceName)
}
