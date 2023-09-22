package main

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	damoclesmanager "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/damocles-manager"
	damoclesworker "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/damocles-worker"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/pvc"
)

func main() {
	plugin.SetupPluginFromStdin(damoclesworker.PluginInfo, Exec)
}

type DepParams struct {
	damoclesworker.Config
	PieceStore    pvc.PvcReturn `json:"PieceStore" jsonschema:"PieceStore" title:"PieceStore" require:"true" description:"piece storage"`
	PersistStores pvc.PvcReturn `json:"PersistStores" jsonschema:"PersistStores" title:"PersistStores" require:"true" description:"persist storage"`

	DamoclesManager damoclesmanager.DamoclesManagerReturn `json:"DamoclesManager" jsonschema:"DamoclesManager" title:"Damocles Manager" description:"damocles manager return" require:"true"`
	MinerAddress    address.Address                       `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true" `
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*damoclesworker.DamoclesWorkerReturn, error) {
	return damoclesworker.DeployFromConfig(ctx, k8sEnv, damoclesworker.Config{
		BaseConfig: depParams.BaseConfig,
		VConfig: damoclesworker.VConfig{
			SealPaths:          depParams.SealPaths,
			PieceStores:        []string{depParams.PieceStore.Name},
			PersistStores:      []string{depParams.PersistStores.Name},
			DamoclesManagerUrl: depParams.DamoclesManager.SvcEndpoint.ToMultiAddr(),
			UserToken:          depParams.UserToken,
			MinerAddress:       depParams.MinerAddress.String()[2:],
		},
	})
}
