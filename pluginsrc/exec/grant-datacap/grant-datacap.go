package main

import (
	"context"

	"github.com/filecoin-project/go-address"
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
	Name:        "grant-datacap",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "grant datacap from notary in genesis node",
}

type InputCaseParams struct {
	GenesisNodeRet genesisnode.GenesisReturn `json:"genesisNodeRet" jsonschema:"genesisNodeRet" title:"Genesis Node return Params" require:"true" description:"genesis node return params, used to get genesis node info"`
	DcClient       address.Address           `json:"dcClient" jsonschema:"dcClient" title:"DataCap Client" require:"true" description:"address of the datacap client"`
	Amount         uint64                    `json:"amount" jsonschema:"amount" title:"DataCap Amount" default:"100000000000" require:"true" description:"amount of datacap to grant"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params InputCaseParams) error {
	return genesisnode.GrantDatacap(ctx, k8sEnv, params.GenesisNodeRet.InstanceName, params.DcClient.String(), params.Amount)
}
