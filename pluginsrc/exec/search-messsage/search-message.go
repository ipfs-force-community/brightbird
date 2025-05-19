package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"

	sophonmessager "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-messager"
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
	Messager    sophonmessager.SophonMessagerReturn `json:"SophonMessager"  jsonschema:"SophonMessager"  title:"Sophon Messager" require:"true" description:"messager return"`
	SearchValue string                              `json:"searchValue"  jsonschema:"searchValue"  title:"Search Value" require:"true" description:"search value: message Id or cid"`
	SearchType  string                              `json:"SearchType"  jsonschema:"SearchType"  title:"SearchType" default:"id" require:"true" enum:"id,cid" description:"search type: id,cid"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	pods, err := sophonmessager.GetPods(ctx, k8sEnv, params.Messager.InstanceName)
	if err != nil {
		return err
	}
	searchVal := strings.TrimSpace(params.SearchValue)

	var msgSearchCmd string
	switch params.SearchType {
	case "id":
		msgSearchCmd = fmt.Sprintf("./sophon-messager msg search --id %v", searchVal)
	case "cid":
		if _, err := cid.Parse(searchVal); err != nil {
			return fmt.Errorf("invalid cid %s", searchVal)
		}
		msgSearchCmd = fmt.Sprintf("./sophon-messager msg search --cid %v", searchVal)
	default:
		return fmt.Errorf("search type %s not support", params.SearchType)
	}
	log.Infoln("cmd : ", msgSearchCmd)

	res, err := k8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), "/bin/sh", "-c", msgSearchCmd)
	if err != nil {
		return err
	}
	log.Infoln("msg result: \n", string(res))
	return nil
}
