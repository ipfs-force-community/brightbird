package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/hunjixin/brightbird/env"
	venus_wallet "github.com/hunjixin/brightbird/env/impl/venus-wallet"
	"github.com/hunjixin/brightbird/types"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("venus-wallet-dep")
var Info = venus_wallet.PluginInfo

type DepParams struct {
	Params     venus_wallet.Config `optional:"true"`
	K8sEnv     *env.K8sEnvDeployer
	AdminToken types.AdminToken
	Gateway    env.IVenusGatewayDeployer
	VenusAuth  env.IVenusAuthDeployer
	types.AnnotateOut
}

func Exec(ctx context.Context, depParams DepParams) (env.IVenusWalletDeployer, error) {
	userToken := ""
	if len(depParams.Params.UserName) > 0 {
		endpoint := depParams.VenusAuth.SvcEndpoint()
		if env.Debug {
			var err error
			endpoint, err = depParams.K8sEnv.PortForwardPod(ctx, depParams.VenusAuth.Pods()[0].GetName(), int(depParams.VenusAuth.Svc().Spec.Ports[0].Port))
			if err != nil {
				return nil, err
			}
		}
		authAPIClient, err := jwtclient.NewAuthClient(endpoint.ToHttp(), string(depParams.AdminToken))
		if err != nil {
			return nil, err
		}

		has, err := authAPIClient.HasUser(ctx, depParams.Params.UserName)
		if err != nil {
			return nil, err
		}
		if !has {
			if depParams.Params.CreateIfNotExit {
				_, err = authAPIClient.CreateUser(ctx, &auth.CreateUserRequest{
					Name:    depParams.Params.UserName,
					Comment: types.PtrString("auto create"),
					State:   1,
				})
				if err != nil {
					return nil, err
				}
				log.Infof("create user %s successfully", depParams.Params.UserName)
			} else {
				return nil, fmt.Errorf("user %s not exit", depParams.Params.UserName)
			}
		}
		userToken, err = authAPIClient.GenerateToken(ctx, depParams.Params.UserName, "write", "")
		if err != nil {
			return nil, err
		}
		log.Infof("create token %s successfully", userToken)
	}

	depParams.Params.UserToken = userToken
	deployer, err := venus_wallet.DeployerFromConfig(depParams.K8sEnv, venus_wallet.Config{
		GatewayUrl: depParams.Gateway.SvcEndpoint().ToMultiAddr(),
		UserToken:  userToken,
	}, depParams.Params)
	if err != nil {
		return nil, err
	}
	err = deployer.Deploy(ctx)
	if err != nil {
		return nil, err
	}
	return deployer, nil
}
