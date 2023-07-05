package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	dropletclient "github.com/hunjixin/brightbird/pluginsrc/deploy/droplet-client"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/hunjixin/brightbird/pluginsrc/deploy/venus"
	venuswallet "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-wallet"
)

func main() {
	plugin.SetupPluginFromStdin(dropletclient.PluginInfo, Exec)
}

type DepParams struct {
	dropletclient.Config

	Auth        sophonauth.SophonAuthDeployReturn `json:"SophonAuth" description:"sophon auth return"`
	Venus       venus.VenusDeployReturn           `json:"Venus" description:"venus return"`
	VenusWallet venuswallet.VenusWalletReturn     `json:"VenusWallet" description:"wallet return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, depParams DepParams) (*dropletclient.DropletClientDeployReturn, error) {
	return dropletclient.DeployFromConfig(ctx, k8sEnv, dropletclient.Config{
		BaseConfig: depParams.BaseConfig,
		VConfig: dropletclient.VConfig{
			NodeUrl:     depParams.Venus.SvcEndpoint.ToMultiAddr(),
			UserToken:   depParams.UserToken,
			WalletUrl:   depParams.VenusWallet.SvcEndpoint.ToMultiAddr(),
			WalletToken: depParams.WalletToken,
		},
	})
}
