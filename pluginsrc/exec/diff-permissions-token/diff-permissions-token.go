package main

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	chain "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	types2 "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs-force-community/sophon-auth/auth"
	"github.com/ipfs-force-community/sophon-auth/jwtclient"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/utils"
	"github.com/ipfs-force-community/brightbird/pkg/version"

	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
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
	Auth       sophonauth.SophonAuthDeployReturn `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Venus      venus.VenusDeployReturn           `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	Permission string                            `json:"permission" jsonschema:"permission" title:"Permission" require:"true" description:""`
	// Wallet     address.Address                   `json:"wallet" jsonschema:"wallet" title:"Wallet" require:"true" description:""`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	authAPIClient, err := jwtclient.NewAuthClient(params.Auth.SvcEndpoint.ToHTTP(), params.Auth.AdminToken)
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
	fmt.Println(name, token)

	tokenList, err := authAPIClient.Tokens(ctx, 0, 10)
	if err != nil {
		return err
	}
	fmt.Println("-------list-token-------")
	for _, token := range tokenList {
		if token.Name == name {
			fmt.Printf("%+v\n", *token)
		}
	}

	permission, err := checkPermission(ctx, token, params)
	if err != nil {
		return err
	}

	if !strings.Contains(permission, params.Permission) {
		return fmt.Errorf("token:%v api-check:[%v], not pass", params.Permission, permission)
	}
	fmt.Printf("token:%v api-check:[%v] , pass\n", params.Permission, permission)

	return nil
}

func checkPermission(ctx context.Context, token string, params TestCaseParams) (string, error) {
	chainRPC, closer, err := chain.DialFullNodeRPC(ctx, params.Venus.SvcEndpoint.ToMultiAddr(), token, nil)
	if err != nil {
		return "", err
	}
	defer closer()

	read, write, sign, admin := false, false, false, false

	chainHead, err := chainRPC.ChainHead(ctx)
	read = err == nil && chainHead != nil

	// writeErr := chainRPC.MpoolPublishByAddr(ctx, walletAddr)
	walletAddr, writeErr := chainRPC.WalletNewAddress(ctx, address.BLS)
	write = writeErr == nil
	if write {
		// check sign
		signedMsg, signErr := chainRPC.WalletSignMessage(ctx, walletAddr, &types2.Message{
			From:       walletAddr,
			To:         walletAddr,
			Value:      abi.NewTokenAmount(0),
			GasFeeCap:  abi.NewTokenAmount(0),
			GasPremium: abi.NewTokenAmount(0),
		})
		sign = signErr == nil && signedMsg != nil
		// check admin
		adminAddrs := chainRPC.WalletAddresses(ctx)
		admin = len(adminAddrs) > 0
	}

	resStr := ""
	if read {
		resStr += "read "
	}
	if write {
		resStr += "write "
	}
	if sign {
		resStr += "sign "
	}
	if admin {
		resStr += "admin "
	}

	return resStr, nil
}

// 生成随机数后缀的函数
func generateRandomSuffix() string {
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))

	randomNum := rng.Intn(10000)
	return "_" + strconv.Itoa(randomNum)
}
