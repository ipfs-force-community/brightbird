package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/venus-shared/api"
	chain "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	sophonminer "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-miner"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
	miner "github.com/ipfs-force-community/sophon-miner/api/client"
)

const timeFormat = "2006-01-02T15:04:05"

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "check_winning_post_count",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "check winning post count",
}

type TestCaseParams struct {
	Auth         sophonauth.SophonAuthDeployReturn   `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Venus        venus.VenusDeployReturn             `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	Miner        sophonminer.SophonMinerDeployReturn `json:"SophonMiner"  jsonschema:"SophonMiner" title:"Sophon Miner" description:"sophon miner return" require:"true"`
	MinerAddress address.Address                     `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true"`
	CheckCount   int                                 `json:"checkCount"  jsonschema:"checkCount" title:"CheckCount" require:"true"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	ainfo := api.NewAPIInfo(params.Miner.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken)
	endpoint, err := ainfo.DialArgs("v0")
	if err != nil {
		return fmt.Errorf("get dial args: %v", err)
	}
	requestHeader := http.Header{}
	ainfo.SetAuthHeader(requestHeader)

	client, closer, err := miner.NewMinerRPC(ctx, endpoint, requestHeader)
	if err != nil {
		return fmt.Errorf("miner rpc, addr: %s, error: %v", params.Miner.SvcEndpoint.ToMultiAddr(), err)
	}
	defer closer()

	chainRPC, closer, err := chain.DialFullNodeRPC(ctx, params.Venus.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		return fmt.Errorf("node rpc, addr: %s, token: %s, error: %v", params.Venus.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, err)
	}
	defer closer()

	head, err := chainRPC.ChainHead(ctx)
	if err != nil {
		return err
	}

	networkParams, err := chainRPC.StateGetNetworkParams(ctx)
	if err != nil {
		return nil
	}
	blockDelaySecs := int(networkParams.BlockDelaySecs)

	startHeight := int(head.Height())
	endHeight := startHeight + params.CheckCount

	fmt.Printf("start height: %d, end height: %d, check count: %d\n", startHeight, endHeight, params.CheckCount)

	wait := blockDelaySecs * (params.CheckCount + 1)
	fmt.Printf("%s start wait mined block, need: %d's'\n", time.Now().Format(timeFormat), wait)

	// 等待 miner 出块
	time.Sleep(time.Duration(wait) * time.Second)

	minedBlock, err := getMinedBlock(ctx, params.MinerAddress, chainRPC, startHeight, endHeight)
	if err != nil {
		return err
	}
	fmt.Printf("miner %s mined %d blocks: %v\n", params.MinerAddress, len(minedBlock), minedBlock)

	// call api to compute winner count
	winners := make(map[abi.ChainEpoch]struct{})
	res, err := client.CountWinners(ctx, []address.Address{params.MinerAddress}, abi.ChainEpoch(startHeight)-1, abi.ChainEpoch(endHeight)-1)
	if err != nil {
		return nil
	}
	for _, one := range res[0].WinEpochList {
		winners[one.Epoch] = struct{}{}
	}
	fmt.Printf("miner %s expect mined %d blocks: %v\n", params.MinerAddress, len(winners), winners)

	if len(minedBlock) != len(winners) {
		msg := fmt.Sprintf("block number not match, expect: %d, actual: %d", len(winners), len(minedBlock))
		fmt.Println(msg)
		return fmt.Errorf(msg)
	}
	for h := range winners {
		if _, ok := minedBlock[h]; !ok {
			msg := fmt.Sprintf("block not found at: %d", h)
			fmt.Println(msg)
			return fmt.Errorf(msg)
		}
	}

	ts, err := chainRPC.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(endHeight), vTypes.EmptyTSK)
	if err != nil {
		return err
	}

	return checkActualComputeCount("log.txt", ts.MinTimestamp(), params.CheckCount, blockDelaySecs)
}

func getMinedBlock(ctx context.Context, miner address.Address, chainRPC chain.FullNode, startHeight, endHeight int) (map[abi.ChainEpoch]uint64, error) {
	ts, err := chainRPC.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(endHeight), vTypes.EmptyTSK)
	if err != nil {
		return nil, err
	}
	blocks := make(map[abi.ChainEpoch]uint64, 0)

	for ts.Height() >= abi.ChainEpoch(startHeight) {
		for _, blk := range ts.Blocks() {
			if blk.Miner == miner {
				blocks[ts.Height()] = blk.Timestamp
				break
			}
		}
		ts, err = chainRPC.ChainGetTipSet(ctx, ts.Parents())
		if err != nil {
			return nil, err
		}
	}

	return blocks, nil
}

var (
	// 2023-09-22T09:56:42.217+0800    INFO    miner   miner/multiminer.go:320 mining compute  {"number of wins": 0, "total miner": 7}
	logStr = "mining compute"
	// 2023-09-22T09:57:42.000+0800    INFO    miner   miner/multiminer.go:296 sync status     {"HeightDiff": 0, "err:": null}
	logStr2 = "sync status"
)

// 通过日志检查实际计算的出块权的次数
func checkActualComputeCount(logFile string, minedBlockTime uint64, checkCount, blockDelaySecs int) error {
	file, err := os.Open(logFile)
	if err != nil {
		return err
	}
	defer file.Close() //nolint

	var expectMinedCount, minedCount int
	checkPoints := make(map[string]struct{}, checkCount)
	lastCheckTime := time.Unix(int64(minedBlockTime+uint64(2*blockDelaySecs)), 0)

	for i := 0; i < checkCount; i++ {
		t := time.Unix(int64(minedBlockTime-uint64(i*blockDelaySecs)), 0).UTC()
		str := t.Format(timeFormat)
		str = str[:len(str)-3]
		if i == checkCount-1 && t.Second() >= 30 {
			expectMinedCount++
		}
		expectMinedCount++
		checkPoints[str] = struct{}{}
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, logStr2) {
			logTime, err := getLogTime(line)
			if err != nil {
				return err
			}
			fmt.Printf("%s: %s\n", logStr2, logTime.String())
			if logTime.After(lastCheckTime) {
				return fmt.Errorf("log count: %s, %s, %d < %d", logTime, lastCheckTime, minedCount, expectMinedCount)
			}
		}

		if !strings.Contains(line, logStr) {
			continue
		}

		logTime, err := getLogTime(line)
		if err != nil {
			return err
		}
		str := logTime.Format(timeFormat)
		str = str[:len(str)-3]
		if _, ok := checkPoints[str]; ok {
			fmt.Println("expected log: ", line)
			minedCount++
		}
		if minedCount == expectMinedCount {
			fmt.Printf("check actual compute count success: %d\n", expectMinedCount)
			break
		}
	}

	return scanner.Err()
}

func getLogTime(log string) (time.Time, error) {
	cols := strings.Split(log, "\t")
	if len(cols) < 2 {
		return time.Time{}, fmt.Errorf("invalid log: %s", log)
	}
	// 2023-09-16T07:21:53.038+0800 => 2023-09-16T07:21:53
	simpleTimeStr := strings.Split(cols[0], ".")[0]
	logTime, err := time.Parse(timeFormat, simpleTimeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse %s failed: %v", cols[0], err)
	}

	return logTime, nil
}
