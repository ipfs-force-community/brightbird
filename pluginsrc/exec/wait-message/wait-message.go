package main

import (
	"context"
	"strings"

	"github.com/filecoin-project/venus/venus-shared/api/messager"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	sophonmessager "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-messager"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("wait-message")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "wait-message",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "wait a messager msg id for result",
}

type TestCaseParams struct {
	Auth       sophonauth.SophonAuthDeployReturn   `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Messager   sophonmessager.SophonMessagerReturn `json:"SophonMessager"  jsonschema:"SophonMessager"  title:"Sophon Messager" require:"true" description:"messager return"`
	MessageId  string                              `json:"MessageId"  jsonschema:"MessageId"  title:"MessageId" require:"true" description:"MessageId"`
	Confidence int                                 `json:"confidence"  jsonschema:"confidence"  title:"Confidence" default:"5" require:"true" description:"confience height for wait message"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	messagerRPC, closer, err := messager.DialIMessagerRPC(ctx, params.Messager.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	msg, err := messagerRPC.WaitMessage(ctx, strings.TrimSpace(params.MessageId), uint64(params.Confidence))
	if err != nil {
		return err
	}

	if msg.Receipt.ExitCode != 0 {
		log.Errorln("message fail %d", msg.Receipt.ExitCode)
		return err
	}

	log.Infoln("message cid ", msg.SignedCid)
	log.Infoln("Height:", msg.Height)
	log.Infoln("Tipset:", msg.TipSetKey.String())
	log.Infoln("exitcode:", msg.Receipt.ExitCode)
	log.Infoln("gas_used:", msg.Receipt.GasUsed)
	log.Infoln("return_value:", string(msg.Receipt.Return))
	return nil
}
