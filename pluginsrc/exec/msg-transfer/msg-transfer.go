package main

import (
	"context"
	"time"

	"github.com/filecoin-project/go-address"
	fbig "github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	vtypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"
)

var log = logging.Logger("msg-transfer")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "msg-transfer",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "transfer through messager",
}

type TestCaseParams struct {
	fx.In
	Params struct {
		Amount string `json:"amount" description:"specify amount of fil to transfer"`
	} `optional:"true"`

	K8sEnv        *env.K8sEnvDeployer `json:"-"`
	SophonMessage env.IDeployer       `svcname:"SophonMessage"`
	From          env.IExec           `svcname:"From"`
	FromToken     env.IExec           `svcname:"FromToken"`
	To            env.IExec           `svcname:"To"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	endpoint, err := params.SophonMessage.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	if env.Debug {
		pods, err := params.SophonMessage.Pods(ctx)
		if err != nil {
			return nil, err
		}

		svc, err := params.SophonMessage.Svc(ctx)
		if err != nil {
			return nil, err
		}
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return nil, err
		}
	}

	tokenP, err := params.FromToken.Param("Token")
	if err != nil {
		return nil, err
	}

	token, err := env.UnmarshalJSON[string](tokenP.Raw())
	if err != nil {
		return nil, err
	}

	messagerClient, closer, err := messager.DialIMessagerRPC(ctx, endpoint.ToMultiAddr(), token, nil)
	if err != nil {
		return nil, err
	}
	defer closer()

	fromAddrP, err := params.From.Param("ImportAddr") //todo change to dynamic value
	if err != nil {
		return nil, err
	}
	fromAddr, err := env.UnmarshalJSON[address.Address](fromAddrP.Raw())
	if err != nil {
		return nil, err
	}

	toAddrP, err := params.To.Param("Wallet")
	if err != nil {
		return nil, err
	}
	toAddr, err := env.UnmarshalJSON[address.Address](toAddrP.Raw())
	if err != nil {
		return nil, err
	}

	fil, err := vtypes.ParseFIL(params.Params.Amount)
	if err != nil {
		return nil, err
	}

	uid, err := messagerClient.PushMessage(ctx, &vtypes.Message{
		Version: 0,
		To:      toAddr,
		From:    fromAddr,
		Nonce:   0,
		Value:   fbig.Int(fil),
		Method:  0,
	}, nil)
	if err != nil {
		return nil, err
	}

	log.Infof("send message and get uid %s", uid)

	ctx, cancel := context.WithTimeout(ctx, time.Minute*10)
	defer cancel()
	_, err = messagerClient.WaitMessage(ctx, uid, 3)
	if err != nil {
		return nil, err
	}

	return env.NewSimpleExec(), nil
}
