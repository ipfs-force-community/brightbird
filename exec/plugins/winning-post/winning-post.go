package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/filecoin-project/go-address"
	miner "github.com/filecoin-project/venus-miner/api/client"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/filecoin-project/venus/venus-shared/api/wallet"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
	"math/rand"
	"os"
	"regexp"
)

var Info = types.PluginInfo{
	Name:        "add_miner",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "user add miner",
}

type TestCaseParams struct {
	fx.In
	AdminToken                 types.AdminToken
	K8sEnv                     *env.K8sEnvDeployer             `json:"-"`
	VenusWallet                env.IVenusWalletDeployer        `json:"-" svcname:"Wallet"`
	VenusMiner                 env.IVenusMinerDeployer         `json:"-"`
	VenusMessage               env.IVenusMessageDeployer       `json:"-"`
	VenusSectorManagerDeployer env.IVenusSectorManagerDeployer `json:"-"`
}

func Exec(ctx context.Context, params TestCaseParams) error {
	walletAddr, err := CreateWallet(ctx, params)
	if err != nil {
		fmt.Printf("create wallet failed: %v\n", err)
		return err
	}

	minerAddr, err := CreateMiner(ctx, params, walletAddr)
	if err != nil {
		fmt.Printf("create miner failed: %v\n", err)
		return err
	}

	minerInfo, err := GetMinerInfo(ctx, params, minerAddr)
	if err != nil {
		fmt.Printf("get miner info failed: %v\n", err)
		return err
	}
	fmt.Println("miner info: %v", minerInfo)

	getMiner, err := GetMinerFromVenusMiner(ctx, params, minerAddr)
	if err != nil {
		fmt.Printf("get miner for venus_miner failed: %v\n", err)
	}
	fmt.Println("miner info: %v", getMiner)

	WinningPostMsg, err := GetWinningPostMsg(ctx, params, minerAddr)
	if err != nil {
		fmt.Printf("get miner for venus_miner failed: %v\n", err)
	}
	fmt.Println("winning post message is: %v", WinningPostMsg)

	return nil
}

func CreateWallet(ctx context.Context, params TestCaseParams) (address.Address, error) {
	walletToken, err := env.ReadWalletToken(ctx, params.K8sEnv, params.VenusWallet.Pods()[0].GetName())
	if err != nil {
		return address.Undef, fmt.Errorf("read wallet token failed: %w\n", err)
	}

	endpoint := params.VenusWallet.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, params.VenusWallet.Pods()[0].GetName(), int(params.VenusWallet.Svc().Spec.Ports[0].Port))
		if err != nil {
			return address.Undef, fmt.Errorf("port forward failed: %w\n", err)
		}
	}

	walletRpc, closer, err := wallet.DialIFullAPIRPC(ctx, endpoint.ToMultiAddr(), walletToken, nil)
	if err != nil {
		return address.Undef, fmt.Errorf("dial iFullAPI rpc failed: %w\n", err)
	}
	defer closer()

	password := "123456"
	err = walletRpc.SetPassword(ctx, password)
	if err != nil {
		return address.Undef, fmt.Errorf("set password failed: %w\n", err)
	}

	walletAddr, err := walletRpc.WalletNew(ctx, vTypes.KTBLS)
	if err != nil {
		return address.Undef, fmt.Errorf("create wallet failed: %w\n", err)
	}
	fmt.Printf("wallet: %v\n", walletAddr)

	return walletAddr, nil
}

func CreateMiner(ctx context.Context, params TestCaseParams, walletAddr address.Address) (string, error) {
	cmd := []string{
		"./venus-sector-manager",
		"util",
		"miner",
		"create",
		"--sector-size=8MiB",
		"--exid=" + string(rune(rand.Intn(1000000))),
	}
	cmd = append(cmd, "--from="+walletAddr.String())

	minerAddr, err := params.K8sEnv.ExecRemoteCmd(ctx, params.VenusSectorManagerDeployer.Pods()[0].GetName(), cmd)
	if err != nil {
		return "", fmt.Errorf("exec remote cmd failed: %w\n", err)
	}

	return string(minerAddr), nil
}

func GetMinerInfo(ctx context.Context, params TestCaseParams, minerAddr string) (string, error) {
	getMinerCmd := []string{
		"./venus-sector-manager",
		"util",
		"miner",
		"info",
		minerAddr,
	}
	minerInfo, err := params.K8sEnv.ExecRemoteCmd(ctx, params.VenusSectorManagerDeployer.Pods()[0].GetName(), getMinerCmd)
	if err != nil {
		return "", fmt.Errorf("exec remote cmd failed: %w\n", err)
	}

	return string(minerInfo), nil
}

func GetMinerFromVenusMiner(ctx context.Context, params TestCaseParams, minerAddr string) (string, error) {

	endpoint := params.VenusMiner.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, params.VenusMiner.Pods()[0].GetName(), int(params.VenusMiner.Svc().Spec.Ports[0].Port))
		if err != nil {
			return "", err
		}
	}

	client, closer, err := miner.NewMinerRPC(ctx, endpoint.ToHttp(), nil)
	if err != nil {
		return "", err
	}
	defer closer()

	list, err := client.ListAddress(ctx)
	for _, m := range list {
		// 使用 miner 进行操作
		if m.Id == minerAddr {
			return minerAddr, nil
		}
	}

	return "", nil
}

func GetWinningPostMsg(ctx context.Context, params TestCaseParams, authToken string) (string, error) {
	endpoint := params.VenusMessage.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, params.VenusMessage.Pods()[0].GetName(), int(params.VenusMessage.Svc().Spec.Ports[0].Port))
		if err != nil {
			return "", err
		}
	}

	client, closer, err := messager.DialIMessagerRPC(ctx, endpoint.ToHttp(), authToken, nil)
	if err != nil {
		return "", err
	}
	defer closer()

	// Get message IDs.
	wdpostID, err := readLogForMsgIds()
	if err != nil {
		return "", fmt.Errorf("failed to get message IDs: %v", err)
	}

	_, err = client.GetMessageByUid(ctx, wdpostID)
	if err != nil {
		return "", err
	}

	return "", nil
}

func readLogForMsgIds() (string, error) {
	file, err := os.Open("log.txt")
	if err != nil {
		return "", err
	}
	defer file.Close()

	var msgWdpostId string

	reWdpost := regexp.MustCompile(`Submitted window post: (\w+)`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if match := reWdpost.FindStringSubmatch(line); len(match) > 0 {
			msgWdpostId = match[1]
		}
		if msgWdpostId != "" {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	fmt.Printf("msg_wdpost_id: %v\n", msgWdpostId)
	return msgWdpostId, nil
}