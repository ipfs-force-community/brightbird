package main

import (
	"context"
	"fmt"

	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"

	sophonmessager "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-messager"
)

var log = logging.Logger("msg-list")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "sophon-messager-msg-list",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "sophon-messager msg list",
}

type TestCaseParams struct {
	Messager sophonmessager.SophonMessagerReturn `json:"SophonMessager"  jsonschema:"SophonMessager"  title:"Sophon Messager" require:"true" description:"messager return"`
	State    int                                 `json:"state"  jsonschema:"state"  title:"State" default:"1" require:"true" description:"message state:{1:  UnFillMsg 2:  FillMsg 3:  OnChainMsg 4:  FailedMsg 5:  NonceConflictMsg 6:  NoWalletMsg}"` // state: 1:  UnFillMsg 2:  FillMsg 3:  OnChainMsg 4:  FailedMsg 5:  NonceConflictMsg 6:  NoWalletMsg
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	pods, err := sophonmessager.GetPods(ctx, k8sEnv, params.Messager.InstanceName)
	if err != nil {
		return err
	}
	msgListCmd := fmt.Sprintf("./sophon-messager msg list --state %v", params.State)
	log.Infof("msgListCmd : %v", msgListCmd)

	res, err := k8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), "/bin/sh", "-c", msgListCmd)
	if err != nil {
		return err
	}
	log.Infof("result : \n%v", string(res))

	return nil
}
