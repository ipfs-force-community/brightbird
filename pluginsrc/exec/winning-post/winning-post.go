package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	damoclesmanager "github.com/hunjixin/brightbird/pluginsrc/deploy/damocles-manager"
	sophonmessager "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-messager"
	sophonminer "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-miner"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	miner "github.com/ipfs-force-community/sophon-miner/api/client"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "winning_post",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "check miner winning post if success.",
}

type TestCaseParams struct {
	SophonMessager  sophonmessager.SophonMessagerReturn   `json:"SophonMessager"`
	DamoclesManager damoclesmanager.DamoclesManagerReturn `json:"DamoclesManager"`
	SophonMiner     sophonminer.SophonMinerDeployReturn   `json:"SophonMiner"`
	MinerAddress    address.Address                       `json:"MinerAddress" type:"string"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	getMiner, err := GetMinerFromVenusMiner(ctx, params, params.MinerAddress.String())
	if err != nil {
		fmt.Printf("get miner for venus_miner failed: %v\n", err)
	}
	fmt.Printf("miner info: %v\n", getMiner)

	WinningPostMsg, err := GetWinningPostMsg(ctx, params, params.MinerAddress.String())
	if err != nil {
		fmt.Printf("get miner for venus_miner failed: %v\n", err)
	}
	fmt.Printf("winning post message is: %v\n", WinningPostMsg)

	return nil
}

func GetMinerFromVenusMiner(ctx context.Context, params TestCaseParams, minerAddr string) (string, error) {
	client, closer, err := miner.NewMinerRPC(ctx, params.SophonMiner.SvcEndpoint.ToMultiAddr(), nil)
	if err != nil {
		return "", err
	}
	defer closer()

	list, err := client.ListAddress(ctx)
	if err != nil {
		return "", err
	}

	for _, m := range list {
		// 使用 miner 进行操作
		if m.Id == minerAddr {
			return minerAddr, nil
		}
	}

	return "", nil
}

func GetWinningPostMsg(ctx context.Context, params TestCaseParams, authToken string) (string, error) {
	client, closer, err := messager.DialIMessagerRPC(ctx, params.SophonMessager.SvcEndpoint.ToMultiAddr(), authToken, nil)
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
	defer file.Close() //nolint

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
