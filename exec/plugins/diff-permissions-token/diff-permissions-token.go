package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/jwtclient"
	chain "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/filecoin-project/venus/venus-shared/api/wallet"
	types2 "github.com/filecoin-project/venus/venus-shared/types"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "admin/sign/write/read token ",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "generate diff permissions token",
}

type TestCaseParams struct {
	fx.In
	Params struct {
		Permission string `json:"permission"`
	} `optional:"true"`
	AdminToken  types.AdminToken
	K8sEnv      *env.K8sEnvDeployer      `json:"-"`
	VenusAuth   env.IVenusAuthDeployer   `json:"-"`
	Venus       env.IVenusDeployer       `json:"-"`
	VenusWallet env.IVenusWalletDeployer `json:"-"`
}

func Exec(ctx context.Context, params TestCaseParams) error {
	endpoint := params.VenusAuth.SvcEndpoint()
	if env.Debug {
		venusAuthPods, err := params.VenusAuth.Pods(ctx)
		if err != nil {
			return err
		}

		svc, err := params.VenusAuth.Svc(ctx)
		if err != nil {
			return err
		}
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, venusAuthPods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return err
		}
	}
	authAPIClient, err := jwtclient.NewAuthClient(endpoint.ToHttp(), string(params.AdminToken))
	if err != nil {
		return err
	}

	_, err = authAPIClient.CreateUser(ctx, &auth.CreateUserRequest{
		Name:    params.Params.Permission,
		Comment: utils.StringPtr("comment " + params.Params.Permission),
		State:   0,
	})
	if err != nil {
		return err
	}

	token, err := authAPIClient.GenerateToken(ctx, params.Params.Permission, params.Params.Permission, "")
	if err != nil {
		return err
	}
	fmt.Println(token)

	permission, err := checkPermission(ctx, token, params)
	if err != nil {
		return err
	}
	if permission != params.Params.Permission {
		return err
	}
	return nil
}

func checkPermission(ctx context.Context, token string, params TestCaseParams) (string, error) {
	endpoint := params.Venus.SvcEndpoint()
	if env.Debug {
		venusPods, err := params.Venus.Pods(ctx)
		if err != nil {
			return "", err
		}

		svc, err := params.Venus.Svc(ctx)
		if err != nil {
			return "", err
		}
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, venusPods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return "", err
		}
	}
	chainRpc, closer, err := chain.DialFullNodeRPC(ctx, endpoint.ToMultiAddr(), token, nil)
	if err != nil {
		return "", err
	}
	defer closer()

	walletAddr, err := createWallet(ctx, params)
	if err != nil {
		return "", err
	}

	chainHead, err := chainRpc.ChainHead(ctx)
	read := err == nil && chainHead != nil

	writeErr := chainRpc.MpoolPublishByAddr(ctx, walletAddr)
	write := writeErr == nil

	msg := types2.Message{
		From:       walletAddr,
		To:         walletAddr,
		Value:      abi.NewTokenAmount(0),
		GasFeeCap:  abi.NewTokenAmount(0),
		GasPremium: abi.NewTokenAmount(0),
	}

	signedMsg, signErr := chainRpc.WalletSignMessage(ctx, walletAddr, &msg)
	sign := signErr == nil && signedMsg != nil

	adminAddrs := chainRpc.WalletAddresses(ctx)
	admin := adminAddrs != nil && len(adminAddrs) > 0

	if read && !write && !sign && !admin {
		return "read", nil
	}
	if !read && write && !sign && !admin {
		return "write", nil
	}
	if !read && !write && sign && !admin {
		return "sign", nil
	}
	if !read && !write && !sign && admin {
		return "admin", nil
	}

	return "", nil
}

func createWallet(ctx context.Context, params TestCaseParams) (address.Address, error) {
	venusWalletPods, err := params.VenusWallet.Pods(ctx)
	if err != nil {
		return address.Undef, err
	}

	svc, err := params.Venus.Svc(ctx)
	if err != nil {
		return address.Undef, err
	}
	walletToken, err := env.ReadWalletToken(ctx, params.K8sEnv, venusWalletPods[0].GetName())
	if err != nil {
		return address.Undef, err
	}

	endpoint := params.VenusWallet.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, venusWalletPods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return address.Undef, err
		}
	}

	walletRpc, closer, err := wallet.DialIFullAPIRPC(ctx, endpoint.ToMultiAddr(), walletToken, nil)
	if err != nil {
		return address.Undef, err
	}
	defer closer()

	password := "123456"
	err = walletRpc.SetPassword(ctx, password)
	if err != nil {
		return address.Undef, err
	}
	walletAddr, err := walletRpc.WalletNew(ctx, vTypes.KTBLS)
	if err != nil {
		return address.Undef, err
	}
	fmt.Println("wallet:", walletAddr)
	return walletAddr, nil
}
