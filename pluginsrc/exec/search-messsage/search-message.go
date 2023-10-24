package main

import (
	"context"
	"strings"

	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	sophonmessager "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-messager"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("search-messsage")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "search-messsage",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "search message",
}

type TestCaseParams struct {
	Messager   sophonmessager.SophonMessagerReturn `json:"SophonMessager"  jsonschema:"SophonMessager"  title:"Sophon Messager" require:"true" description:"messager return"`
	MessageId  string                              `json:"MessageId"  jsonschema:"MessageId"  title:"MessageId" require:"true" description:"MessageId"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	// ./sophon-messager msg search --id bafk4bzacialvjurmovbzow7livpluvvhym6owrhirpugfzpu3hwysblguwoian5f45kyxvyoylkq6lljd2d76mzz2thl7arfb43emp7hcsez5h3x
	pods, err := sophonmessager.GetPods(ctx, k8sEnv, params.Messager.InstanceName)
	if err != nil {
		return err
	}
	msgListCmd := "./sophon-messager msg list --state 1"
	log.Infoln("msgListCmd is: ", msgListCmd)

	res, err := k8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), "/bin/sh", "-c", msgListCmd)
	if err != nil {
		return err
	}
	log.Infoln("msg list result is: ", string(res))

	msgSearchCmd := "./sophon-messager msg search --id " + strings.TrimSpace(params.MessageId)
	log.Infoln("upsertCmd is: ", msgSearchCmd)

	res, err = k8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), "/bin/sh", "-c", msgSearchCmd)
	if err != nil {
		return err
	}
	log.Infoln("msg result is: ", string(res))
	return nil
}
