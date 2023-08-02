package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	sophonmessager "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-messager"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "message_log",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "message log",
}

type TestCaseParams struct {
	Auth           sophonauth.SophonAuthDeployReturn   `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	SophonMessager sophonmessager.SophonMessagerReturn `json:"SophonMessager"  jsonschema:"SophonMessager"  title:"Sophon Messager" require:"true" description:"messager return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	client, closer, err := messager.DialIMessagerRPC(ctx, params.SophonMessager.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	// Get message IDs.
	precommitID, procommitID, wdpostID, faultID, err := readLogForMsgIds()
	if err != nil {
		return fmt.Errorf("failed to get message IDs: %v", err)
	}

	_, err = client.GetMessageByUid(ctx, precommitID)
	if err != nil {
		return err
	}

	_, err = client.GetMessageByUid(ctx, procommitID)
	if err != nil {
		return err
	}

	_, err = client.GetMessageByUid(ctx, wdpostID)
	if err != nil {
		return err
	}

	_, err = client.GetMessageByUid(ctx, faultID)
	if err != nil {
		return err
	}

	return nil
}

func readLogForMsgIds() (string, string, string, string, error) {
	file, err := os.Open("log.txt")
	if err != nil {
		return "", "", "", "", err
	}
	defer file.Close() //nolint

	var msgPrecommitId, msgProcommitId, msgWdpostId, msgFaultId string
	rePrecommit := regexp.MustCompile(`"stage": "pre-commit", "msg-cid": "(\w+)"`)
	reProcommit := regexp.MustCompile(`"stage": "prove-commit", "msg-cid": "(\w+)"`)
	reWdpost := regexp.MustCompile(`Submitted window post: (\w+)`)
	reFault := regexp.MustCompile(`declare faults recovered message published\s+{"message-id": "(\w+)"}`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if match := rePrecommit.FindStringSubmatch(line); len(match) > 0 {
			msgPrecommitId = match[1]
		} else if match := reProcommit.FindStringSubmatch(line); len(match) > 0 {
			msgProcommitId = match[1]
		} else if match := reWdpost.FindStringSubmatch(line); len(match) > 0 {
			msgWdpostId = match[1]
		} else if match := reFault.FindStringSubmatch(line); len(match) > 0 {
			msgFaultId = match[1]
		}
		if msgPrecommitId != "" && msgProcommitId != "" && msgWdpostId != "" && msgFaultId != "" {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", "", "", err
	}

	fmt.Printf("msg_precommit_id: %v\nmsg_procommit_id: %v\nmsg_wdpost_id: %v\nmsg_fault_id: %v\n",
		msgPrecommitId, msgProcommitId, msgWdpostId, msgFaultId)
	return msgPrecommitId, msgProcommitId, msgWdpostId, msgFaultId, nil
}
