package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/ipfs-force-community/brightbird/types"

	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	"github.com/ipfs-force-community/brightbird/version"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "faucet",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "used to get money avoid nonce conflict",
}

type TestCaseParams struct {
	RpcUrl      string          `json:"rpcUrl" jsonschema:"rpcUrl" title:"RPC URL(lotus/venus)" description:"lotus or venus rpc url /ip4/x.x.x.x/1234"`
	RPCToken    string          `json:"rpcToken" jsonschema:"rpcToken" title:"RPC Token(lotus/venus)" description:"rpc token"`
	FromAddr    address.Address `json:"fromAddr"  jsonschema:"fromAddr"  title:"From Address" require:"true" description:"address to send fil"`
	ReceiveAddr address.Address `json:"receiveAddr"  jsonschema:"receiveAddr"  title:"Receive Address" require:"true" description:"address to receive fil"`
	Amount      vTypes.FIL      `json:"amount"  jsonschema:"amount"  title:"Value" default:"0fil" require:"true" description:"amount to get unit(fil/attofil)"`
}

func Exec(ctx context.Context, _ *env.K8sEnvDeployer, params TestCaseParams) error {
	fullNode, closer, err := v1.DialFullNodeRPC(ctx, params.RpcUrl, params.RPCToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	msg, err := fullNode.MpoolPushMessage(ctx, &vTypes.Message{
		From:  params.FromAddr,
		To:    params.ReceiveAddr,
		Value: vTypes.BigInt(params.Amount),
	}, nil)
	if err != nil {
		return err
	}

	lookup, err := fullNode.StateWaitMsg(ctx, msg.Cid(), 1, 0, false)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	if lookup.Receipt.ExitCode != 0 {
		return fmt.Errorf("message %s faied reason %s", msg.Cid(), string(lookup.Receipt.Return))
	}
	return nil
}
