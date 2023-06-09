package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"

	"go.uber.org/fx"

	miner "github.com/filecoin-project/venus-miner/api/client"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
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
	fx.In
	K8sEnv                     *env.K8sEnvDeployer `json:"-"`
	VenusMiner                 env.IDeployer       `json:"-" svcname:"VenusMiner"`
	VenusMessage               env.IDeployer       `json:"-" svcname:"VenusMessage"`
	VenusSectorManagerDeployer env.IDeployer       `json:"-" svcname:"VenusSectorManager"`
	CreateMiner                env.IExec           `json:"-" svcname:"CreateMiner"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	minerAddr, err := params.CreateMiner.Param("Miner")
	if err != nil {
		return nil, err
	}

	minerInfo, err := GetMinerInfo(ctx, params, minerAddr.MustString())
	if err != nil {
		fmt.Printf("get miner info failed: %v\n", err)
		return nil, err
	}
	fmt.Printf("miner info: %v\n", minerInfo)

	getMiner, err := GetMinerFromVenusMiner(ctx, params, minerAddr.MustString())
	if err != nil {
		fmt.Printf("get miner for venus_miner failed: %v\n", err)
	}
	fmt.Printf("miner info: %v\n", getMiner)

	WinningPostMsg, err := GetWinningPostMsg(ctx, params, minerAddr.MustString())
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
		return "", fmt.Errorf("exec remote cmd failed: %w", err)
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

	endpoint, err := params.VenusMiner.SvcEndpoint()
	if err != nil {
		return "", err
	}

	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return "", err
		}
	}

	client, closer, err := miner.NewMinerRPC(ctx, endpoint.ToHTTP(), nil)
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
	endpoint, err := params.VenusMessage.SvcEndpoint()
	if err != nil {
		return "", err
	}

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

	client, closer, err := messager.DialIMessagerRPC(ctx, endpoint.ToHTTP(), authToken, nil)
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
