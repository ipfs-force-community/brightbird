package main

import (
	"context"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"

	dropletclient "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-client"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/pvc"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
	venuswallet "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus-wallet"
)

func main() {
	plugin.SetupPluginFromStdin(dropletclient.PluginInfo, Exec)
}

type DepParams struct {
	dropletclient.Config

	PieceStore  pvc.PvcReturn                     `json:"PieceStore" jsonschema:"PieceStore" title:"PieceStore" default:"" require:"false" description:"piece storage"`
	Auth        sophonauth.SophonAuthDeployReturn `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Venus       venus.VenusDeployReturn           `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	VenusWallet venuswallet.VenusWalletReturn     `json:"VenusWallet" jsonschema:"VenusWallet" title:"Venus Wallet" description:"wallet return" require:"true"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*dropletclient.DropletClientDeployReturn, error) {
	ps := []string{}
	if depParams.PieceStore.Name != "" {
		ps = append(ps, depParams.PieceStore.Name)
	}
	return dropletclient.DeployFromConfig(ctx, k8sEnv, dropletclient.Config{
		BaseConfig: depParams.BaseConfig,
		VConfig: dropletclient.VConfig{
			PieceStores:          ps,
			NodeUrl:              depParams.Venus.SvcEndpoint.ToMultiAddr(),
			UserToken:            depParams.UserToken,
			WalletUrl:            depParams.VenusWallet.SvcEndpoint.ToMultiAddr(),
			WalletToken:          depParams.WalletToken,
			DefaultMarketAddress: depParams.DefaultMarketAddress,
		},
	})
}
