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
	Name:        "diff_permissions_token",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "diff permissions token",
}

type TestCaseParams struct {
	fx.In
	Params struct {
		Permission string `json:"permission"`
	} `optional:"true"`
	AdminToken types.AdminToken
	K8sEnv     *env.K8sEnvDeployer    `json:"-"`
	VenusAuth  env.IVenusAuthDeployer `json:"-"`
}

func Exec(ctx context.Context, params TestCaseParams) error {
	endpoint := params.VenusAuth.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, params.VenusAuth.Pods()[0].GetName(), int(params.VenusAuth.Svc().Spec.Ports[0].Port))
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

	permission, err := checkPermission(ctx, endpoint.ToMultiAddr(), token)
	if err != nil {
		return err
	}
	if permission != params.Params.Permission {
		return err
	}
	return nil
}

func checkPermission(ctx context.Context, addr string, token string) (string, error) {
	chainRpc, closer, err := chain.DialFullNodeRPC(ctx, addr, token, nil)
	if err != nil {
		return "", err
	}
	defer closer()

	walletAddr, err := createWallet(ctx, addr, token)
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

func createWallet(ctx context.Context, addr string, token string) (address.Address, error) {

	walletRpc, closer, err := wallet.DialIFullAPIRPC(ctx, addr, token, nil)
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
