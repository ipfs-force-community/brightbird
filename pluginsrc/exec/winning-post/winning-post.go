package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"

	miner "github.com/filecoin-project/venus-miner/api/client"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "winning_post",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "check miner winning post if success.",
}

type TestCaseParams struct {
	fx.In
	K8sEnv                     *env.K8sEnvDeployer `json:"-"`
	VenusMiner                 env.IDeployer       `json:"-" svcname:"VenusMiner"`
	VenusMessage               env.IDeployer       `json:"-" svcname:"VenusMessage"`
	VenusSectorManagerDeployer env.IDeployer       `json:"-" svcname:"VenusSectorManager"`
	CreateMiner                env.IExec           `json:"-" svcname:"CreateMiner"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	minerAddr, err := params.CreateMiner.Param("CreateMiner")
	if err != nil {
		return nil, err
	}

	minerInfo, err := GetMinerInfo(ctx, params, minerAddr.(string))
	if err != nil {
		fmt.Printf("get miner info failed: %v\n", err)
		return nil, err
	}
	fmt.Printf("miner info: %v\n", minerInfo)

	getMiner, err := GetMinerFromVenusMiner(ctx, params, minerAddr.(string))
	if err != nil {
		fmt.Printf("get miner for venus_miner failed: %v\n", err)
	}
	fmt.Printf("miner info: %v\n", getMiner)

	WinningPostMsg, err := GetWinningPostMsg(ctx, params, minerAddr.(string))
	if err != nil {
		fmt.Printf("get miner for venus_miner failed: %v\n", err)
	}
	fmt.Printf("winning post message is: %v\n", WinningPostMsg)

	return env.NewSimpleExec(), nil
}

func GetMinerInfo(ctx context.Context, params TestCaseParams, minerAddr string) (string, error) {
	getMinerCmd := []string{
		"./venus-sector-manager",
		"util",
		"miner",
		"info",
		minerAddr,
	}
	pods, err := params.VenusSectorManagerDeployer.Pods(ctx)
	if err != nil {
		return "", err
	}

	minerInfo, err := params.K8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), getMinerCmd...)
	if err != nil {
		return "", fmt.Errorf("exec remote cmd failed: %w\n", err)
	}

	return string(minerInfo), nil
}

func GetMinerFromVenusMiner(ctx context.Context, params TestCaseParams, minerAddr string) (string, error) {
	pods, err := params.VenusMiner.Pods(ctx)
	if err != nil {
		return "", err
	}

	svc, err := params.VenusMiner.Svc(ctx)
	if err != nil {
		return "", err
	}

	endpoint := params.VenusMiner.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
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
		messagePods, err := params.VenusMessage.Pods(ctx)
		if err != nil {
			return "", err
		}

		svc, err := params.VenusMessage.Svc(ctx)
		if err != nil {
			return "", err
		}

		endpoint, err = params.K8sEnv.PortForwardPod(ctx, messagePods[0].GetName(), int(svc.Spec.Ports[0].Port))
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
