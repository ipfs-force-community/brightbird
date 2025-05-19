package main

import (
	"context"
	"strings"

	"github.com/filecoin-project/venus/venus-shared/api/messager"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"

	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	sophonmessager "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-messager"
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
		log.Errorf("message fail %d", msg.Receipt.ExitCode)
		return err
	}
	log.With(
		"message", msg.SignedCid.String(),
		"Height:", msg.Height,
		"Tipset:", msg.TipSetKey.String(),
		"exitcode:", msg.Receipt.ExitCode,
		"gas_used:", msg.Receipt.GasUsed,
		"return_value:", string(msg.Receipt.Return),
	).Info("wait message success")
	return nil
}
