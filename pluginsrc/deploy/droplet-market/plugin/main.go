package main

import (
	"context"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletmarket "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/pvc"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	sophongateway "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-gateway"
	sophonmessager "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-messager"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
)

var Info = dropletmarket.PluginInfo

func main() {
	plugin.SetupPluginFromStdin(dropletmarket.PluginInfo, Exec)
}

type DepParams struct {
	dropletmarket.Config

	PieceStore pvc.PvcReturn                       `json:"PieceStore" jsonschema:"PieceStore" title:"PieceStore" require:"true" description:"piece storage"`
	Auth       sophonauth.SophonAuthDeployReturn   `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Venus      venus.VenusDeployReturn             `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	Gateway    sophongateway.SophonGatewayReturn   `json:"SophonGateway"  jsonschema:"SophonGateway"  title:"SophonGateway" require:"true" description:"gateway deploy return"`
	Messager   sophonmessager.SophonMessagerReturn `json:"SophonMessager"  jsonschema:"SophonMessager"  title:"Sophon Messager" require:"true" description:"messager return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*dropletmarket.DropletMarketDeployReturn, error) {
	return dropletmarket.DeployFromConfig(ctx, k8sEnv, dropletmarket.Config{
		BaseConfig: depParams.BaseConfig,
		VConfig: dropletmarket.VConfig{
			PieceStores: []string{depParams.PieceStore.Name},
			NodeUrl:     depParams.Venus.SvcEndpoint.ToMultiAddr(),
			GatewayUrl:  depParams.Gateway.SvcEndpoint.ToMultiAddr(),
			MessagerUrl: depParams.Messager.SvcEndpoint.ToMultiAddr(),
			AuthUrl:     depParams.Auth.SvcEndpoint.ToHTTP(),

			UserToken: depParams.UserToken,
			UseMysql:  depParams.UseMysql,
		},
	})
}
