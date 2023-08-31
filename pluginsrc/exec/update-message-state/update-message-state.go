package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/venus/venus-shared/api/messager"
	mTypes "github.com/filecoin-project/venus/venus-shared/types/messager"
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
	Name:        "update_message_state",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "update message state through sophon messager",
}

type TestCaseParams struct {
	Auth     sophonauth.SophonAuthDeployReturn   `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Messager sophonmessager.SophonMessagerReturn `json:"SophonMessager"  jsonschema:"SophonMessager"  title:"Sophon Messager" require:"true" description:"messager return"`

	ID string `json:"id"  jsonschema:"id"  title:"Message'ID" require:"true" description:"messager's ID"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	client, closer, err := messager.DialIMessagerRPC(ctx, params.Messager.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	msg, err := client.GetMessageByUid(ctx, params.ID)
	if err != nil {
		return err
	}
	if msg.State != mTypes.OnChainMsg {
		return fmt.Errorf("message state expect %d, but got %d", mTypes.OnChainMsg, msg.State)
	}

	if err := updateMessageState(ctx, client, params.ID, mTypes.FailedMsg); err != nil {
		return fmt.Errorf("failed update message state to %s, %v", mTypes.FailedMsg, err)
	}

	// revert back to original state
	if err := updateMessageState(ctx, client, params.ID, mTypes.OnChainMsg); err != nil {
		return fmt.Errorf("failed update message state to %s, %v", mTypes.OnChainMsg, err)
	}

	return nil
}

func updateMessageState(ctx context.Context, client messager.IMessager, msgID string, newState mTypes.MessageState) error {
	if err := client.UpdateMessageStateByID(ctx, msgID, newState); err != nil {
		return err
	}

	msg, err := client.GetMessageByUid(ctx, msgID)
	if err != nil {
		return err
	}
	if msg.State != newState {
		return fmt.Errorf("message state expect %d, but got %d", newState, msg.State)
	}

	return nil
}
