package main

import (
	"context"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	damoclesmanager "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/damocles-manager"
	dropletmarket "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/pvc"
	sophongateway "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-gateway"
	sophonmessager "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-messager"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
)

func main() {
	plugin.SetupPluginFromStdin(damoclesmanager.PluginInfo, Exec)
}

type DepParams struct {
	damoclesmanager.Config

	PieceStore    pvc.PvcReturn `json:"PieceStore" jsonschema:"PieceStore" title:"PieceStore" require:"true" description:"piece storage"`
	PersistStores pvc.PvcReturn `json:"PersistStores" jsonschema:"PersistStores" title:"PersistStores" require:"true" description:"persist storage"`

	Venus    venus.VenusDeployReturn             `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	Gateway  sophongateway.SophonGatewayReturn   `json:"SophonGateway"  jsonschema:"SophonGateway"  title:"SophonGateway" require:"true" description:"gateway deploy return"`
	Messager sophonmessager.SophonMessagerReturn `json:"SophonMessager"  jsonschema:"SophonMessager"  title:"Sophon Messager" require:"true" description:"messager return"`

	DropletMarket dropletmarket.DropletMarketDeployReturn `json:"DropletMarket" jsonschema:"DropletMarket" title:"DropletMarket" description:"droplet market return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*damoclesmanager.DamoclesManagerReturn, error) {
	return damoclesmanager.DeployFromConfig(ctx, k8sEnv, damoclesmanager.Config{
		BaseConfig: depParams.BaseConfig,
		VConfig: damoclesmanager.VConfig{
			PieceStores:   []string{depParams.PieceStore.Name},
			PersistStores: []string{depParams.PersistStores.Name},
			NodeUrl:       depParams.Venus.SvcEndpoint.ToMultiAddr(),
			MessagerUrl:   depParams.Messager.SvcEndpoint.ToMultiAddr(),
			MarketUrl:     depParams.DropletMarket.SvcEndpoint.ToMultiAddr(),
			GatewayUrl:    depParams.Gateway.SvcEndpoint.ToMultiAddr(),
			UserToken:     depParams.UserToken,
		},
	})
}
