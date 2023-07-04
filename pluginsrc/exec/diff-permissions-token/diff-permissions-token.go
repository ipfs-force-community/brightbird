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
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
	"github.com/hunjixin/brightbird/version"
	"github.com/ipfs-force-community/sophon-auth/auth"
	"github.com/ipfs-force-community/sophon-auth/jwtclient"
	"go.uber.org/fx"
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
	fx.In
	Params struct {
		Permission string `json:"permission"`
	} `optional:"true"`

	K8sEnv     *env.K8sEnvDeployer `json:"-"`
	SophonAuth env.IDeployer       `json:"-" svcname:"SophonAuth"`
	Venus      env.IDeployer       `json:"-" svcname:"Venus"`
	Wallet     env.IExec           `json:"-" svcname:"Wallet"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	endpoint, err := params.SophonAuth.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	adminToken, err := params.SophonAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	authAPIClient, err := jwtclient.NewAuthClient(endpoint.ToHTTP(), adminToken.MustString())
	if err != nil {
		return nil, err
	}

	suffix := generateRandomSuffix()
	name := params.Params.Permission + suffix
	_, err = authAPIClient.CreateUser(ctx, &auth.CreateUserRequest{
		Name:    name,
		Comment: utils.StringPtr("comment " + name),
		State:   0,
	})
	if err != nil {
		return nil, err
	}

	token, err := authAPIClient.GenerateToken(ctx, name, params.Params.Permission, "")
	if err != nil {
		return nil, err
	}
	fmt.Println(token)

	permission, err := checkPermission(ctx, token, params)
	if err != nil {
		return nil, err
	}
	if permission != params.Params.Permission {
		return nil, err
	}
	return env.NewSimpleExec(), nil
}

func checkPermission(ctx context.Context, token string, params TestCaseParams) (string, error) {
	endpoint, err := params.Venus.SvcEndpoint()
	if err != nil {
		return "", err
	}

	chainRPC, closer, err := chain.DialFullNodeRPC(ctx, endpoint.ToMultiAddr(), token, nil)
	if err != nil {
		return "", err
	}
	defer closer()

	walletAddr, err := params.Wallet.Param("Wallet")
	if err != nil {
		return "", err
	}

	chainHead, err := chainRPC.ChainHead(ctx)
	read := err == nil && chainHead != nil

	addr, err := env.UnmarshalJSON[address.Address](walletAddr.Raw())
	if err != nil {
		panic(err)
	}

	writeErr := chainRPC.MpoolPublishByAddr(ctx, addr)
	write := writeErr == nil

	msg := types2.Message{
		From:       addr,
		To:         addr,
		Value:      abi.NewTokenAmount(0),
		GasFeeCap:  abi.NewTokenAmount(0),
		GasPremium: abi.NewTokenAmount(0),
	}

	signedMsg, signErr := chainRPC.WalletSignMessage(ctx, addr, &msg)
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
