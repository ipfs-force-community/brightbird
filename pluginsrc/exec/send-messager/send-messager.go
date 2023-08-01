package main

import (
	"context"
	"encoding/hex"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/ipfs/go-cid"

	vTypes "github.com/filecoin-project/venus/venus-shared/types"
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
	Name:        "send_message",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "send message through sophon messager",
}

type TestCaseParams struct {
	Auth     sophonauth.SophonAuthDeployReturn   `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Messager sophonmessager.SophonMessagerReturn `json:"SophonMessager"  jsonschema:"SophonMessager"  title:"Sophon Messager" require:"true" description:"messager return"`

	From  address.Address `json:"from"  jsonschema:"from"  title:"Message'From" require:"true" description:"messager's send address"`
	To    address.Address `json:"to"  jsonschema:"to"  title:"Message'To" require:"true" description:"messager's to address"`
	Value vTypes.FIL      `json:"value"  jsonschema:"value"  title:"Value" default:"0fil" require:"true" description:"amount to send unit(fil/attofil)"`

	GasPremium abi.TokenAmount `json:"gasPremium"  jsonschema:"gasPremium"  title:"GasPremium" require:"true" description:"gaspremium for miner tip"`
	GasFeeCap  abi.TokenAmount `json:"gasFeeCap"  jsonschema:"gasFeeCap"  title:"GasFeeCap" require:"true" description:"gasfeecap for meet basefeee requirement"`
	GasLimit   int64           `json:"gasLimit"  jsonschema:"gasLimit"  title:"GasLimit" require:"true" description:"limit for gas usage"`
	Method     abi.MethodNum   `json:"method"  jsonschema:"method"  title:"Method" default:"0" require:"true" description:"which method to call"`
	Params     string          `json:"params"  jsonschema:"params"  title:"Params" require:"true" description:"params for method call (hex)"`

	Confidence        int     `json:"confidence"  jsonschema:"confidence"  title:"Confidence" default:"5" require:"true" description:"confience height for wait message"`
	GasOverEstimation float64 `json:"gasOverEstimation"  jsonschema:"gasOverEstimation" default:"1.25" title:"GasOverEstimation" require:"true" description:"extra rate gaslimit"`
	MaxFee            big.Int `json:"maxfee"  jsonschema:"maxfee"  title:"MaxFee" require:"true" description:"maxfee limit for message"`
	GasOverPremium    float64 `json:"gasOverPremium"  jsonschema:"gasOverPremium"  title:"GasOverPremium" require:"true" description:"extra rate gaspremium"`
}

type SendMessageResult struct {
	ID string

	UnsignedCid *cid.Cid
	SignedCid   *cid.Cid
	Receipt     *vTypes.MessageReceipt
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*SendMessageResult, error) {
	client, closer, err := messager.DialIMessagerRPC(ctx, params.Messager.SvcEndpoint.ToMultiAddr(), params.Auth.AdminToken, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	paramsBytes, err := hex.DecodeString(params.Params)
	if err != nil {
		return nil, err
	}

	messagerId, err := client.PushMessage(ctx, &vTypes.Message{
		To:         params.To,
		From:       params.From,
		Value:      abi.TokenAmount{Int: params.Value.Int},
		GasPremium: params.GasPremium,
		GasFeeCap:  params.GasFeeCap,
		GasLimit:   params.GasLimit,

		Method: params.Method,
		Params: paramsBytes,
	}, &mTypes.SendSpec{
		GasOverEstimation: params.GasOverEstimation,
		MaxFee:            params.MaxFee,
		GasOverPremium:    params.GasOverPremium,
	})
	if err != nil {
		return nil, err
	}

	result, err := client.WaitMessage(ctx, messagerId, uint64(params.Confidence))
	if err != nil {
		return nil, err
	}

	return &SendMessageResult{
		ID:          messagerId,
		UnsignedCid: result.UnsignedCid,
		SignedCid:   result.SignedCid,
		Receipt:     result.Receipt,
	}, nil
}
