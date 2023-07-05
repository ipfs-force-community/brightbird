package main

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	chain "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	types2 "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/hunjixin/brightbird/pluginsrc/deploy/venus"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	"github.com/ipfs-force-community/sophon-auth/auth"
	"github.com/ipfs-force-community/sophon-auth/jwtclient"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "admin-sign-write-read-token",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "generate diff permissions token",
}

type TestCaseParams struct {
	Permission string                            `json:"permission"`
	SophonAuth sophonauth.SophonAuthDeployReturn `json:"SophonAuth"`
	Venus      venus.VenusDeployReturn           `json:"Venus"`
	Wallet     address.Address                   `json:"Wallet"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	authAPIClient, err := jwtclient.NewAuthClient(params.SophonAuth.SvcEndpoint.ToHTTP(), params.SophonAuth.AdminToken)
	if err != nil {
		return err
	}

	suffix := generateRandomSuffix()
	name := params.Permission + suffix
	_, err = authAPIClient.CreateUser(ctx, &auth.CreateUserRequest{
		Name:    name,
		Comment: utils.StringPtr("comment " + name),
		State:   0,
	})
	if err != nil {
		return err
	}

	token, err := authAPIClient.GenerateToken(ctx, name, params.Permission, "")
	if err != nil {
		return err
	}
	fmt.Println(token)

	permission, err := checkPermission(ctx, token, params)
	if err != nil {
		return err
	}
	if permission != params.Permission {
		return err
	}
	return nil
}

func checkPermission(ctx context.Context, token string, params TestCaseParams) (string, error) {
	chainRPC, closer, err := chain.DialFullNodeRPC(ctx, params.Venus.SvcEndpoint.ToMultiAddr(), token, nil)
	if err != nil {
		return "", err
	}
	defer closer()

	chainHead, err := chainRPC.ChainHead(ctx)
	read := err == nil && chainHead != nil

	writeErr := chainRPC.MpoolPublishByAddr(ctx, params.Wallet)
	write := writeErr == nil

	msg := types2.Message{
		From:       params.Wallet,
		To:         params.Wallet,
		Value:      abi.NewTokenAmount(0),
		GasFeeCap:  abi.NewTokenAmount(0),
		GasPremium: abi.NewTokenAmount(0),
	}

	signedMsg, signErr := chainRPC.WalletSignMessage(ctx, params.Wallet, &msg)
	sign := signErr == nil && signedMsg != nil

	adminAddrs := chainRPC.WalletAddresses(ctx)
	admin := len(adminAddrs) > 0

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

// 生成随机数后缀的函数
func generateRandomSuffix() string {
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))

	randomNum := rng.Intn(10000)
	return "_" + strconv.Itoa(randomNum)
}
