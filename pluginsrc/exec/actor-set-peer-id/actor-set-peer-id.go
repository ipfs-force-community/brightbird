package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/builtin"
	"github.com/filecoin-project/venus/venus-shared/actors"
	v1api "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	marketapi "github.com/filecoin-project/venus/venus-shared/api/market/v1"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	vtypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/fx"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "actor-set-peer-id",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "actor set peer-id",
}

type TestCaseParams struct {
	fx.In

	K8sEnv                     *env.K8sEnvDeployer `json:"-"`
	VenusAuth                  env.IDeployer       `json:"-" svcname:"VenusAuth"`
	VenusMarket                env.IDeployer       `json:"-" svcname:"VenusMarket"`
	VenusMiner                 env.IDeployer       `json:"-" svcname:"VenusMiner"`
	VenusSectorManagerDeployer env.IDeployer       `json:"-" svcname:"VenusSectorManager"`
	Venus                      env.IDeployer       `json:"-" svcname:"Venus"`
	VenusMessage               env.IDeployer       `json:"-" svcname:"VenusMessage"`
	CreateMiner                env.IExec           `json:"-" svcname:"CreateMiner"`
	NewAddrsListen             env.IExec           `json:"-" svcname:"NewAddrsListen"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	minerAddr, err := params.CreateMiner.Param("Miner")
	if err != nil {
		return nil, err
	}

	messageId, err := SetActorAddr(ctx, params, minerAddr.String())
	if err != nil {
		fmt.Printf("set actor address failed: %v\n", err)
		return nil, err
	}
	fmt.Printf("set actor address message id is: %v\n", messageId)

	err = VertifyMessageIfVaild(ctx, params, messageId)
	if err != nil {
		fmt.Printf("set actor address failed: %v\n", err)
		return nil, err
	}

	return env.NewSimpleExec(), nil
}

func VertifyMessageIfVaild(ctx context.Context, params TestCaseParams, messageId cid.Cid) error {
	adminTokenV, err := params.VenusAuth.Param("AdminToken")
	if err != nil {
		return err
	}

	endpoint, err := params.VenusMessage.SvcEndpoint()
	if err != nil {
		return err
	}
	if env.Debug {
		pods, err := params.VenusMessage.Pods(ctx)
		if err != nil {
			return err
		}

		svc, err := params.VenusMessage.Svc(ctx)
		if err != nil {
			return err
		}
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return err
		}
	}

	client, closer, err := messager.DialIMessagerRPC(ctx, endpoint.ToHttp(), adminTokenV.String(), nil)
	if err != nil {
		return err
	}
	defer closer()

	msg, err := client.GetMessageBySignedCid(ctx, messageId)
	if err != nil {
		return err
	}
	fmt.Printf("Message: %v\n", msg)

	return nil
}

func SetActorAddr(ctx context.Context, params TestCaseParams, minerAddr string) (cid.Cid, error) {
	endpoint, err := params.VenusMarket.SvcEndpoint()
	if err != nil {
		return cid.Undef, err
	}
	if env.Debug {
		pods, err := params.VenusMarket.Pods(ctx)
		if err != nil {
			return cid.Undef, err
		}

		svc, err := params.VenusMarket.Svc(ctx)
		if err != nil {
			return cid.Undef, err
		}

		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return cid.Undef, err
		}
	}
	client, closer, err := marketapi.NewIMarketRPC(ctx, endpoint.ToHttp(), nil)
	if err != nil {
		return cid.Undef, err
	}
	defer closer()

	addrs, err := client.NetAddrsListen(ctx)
	if err != nil && addrs.Addrs != nil {
		return cid.Undef, nil
	}

	pid := addrs.ID

	MessageParams, err := ConstructParams(pid)
	if err != nil {
		return cid.Undef, err
	}

	maddr, err := address.NewFromString(minerAddr)
	if err != nil {
		return cid.Undef, nil
	}

	minfo, err := GetMinerInfo(ctx, params, maddr)
	if err != nil {
		return cid.Undef, err
	}

	mid, err := client.MessagerPushMessage(ctx, &vtypes.Message{
		To:       maddr,
		From:     minfo.Worker,
		Value:    vtypes.NewInt(0),
		GasLimit: 0,
		Method:   builtin.MethodsMiner.ChangeMultiaddrs,
		Params:   MessageParams,
	}, nil)
	if err != nil {
		return cid.Undef, err
	}

	fmt.Printf("Requested multiaddrs change in message %s\n", mid)

	return cid.Undef, err
}

func ConstructParams(pid peer.ID) (param []byte, err error) {

	params, err := actors.SerializeParams(&vtypes.ChangePeerIDParams{NewID: abi.PeerID(pid)})
	if err != nil {
		return nil, err
	}
	return params, nil
}

func GetMinerInfo(ctx context.Context, params TestCaseParams, maddr address.Address) (vtypes.MinerInfo, error) {
	endpoint, err := params.Venus.SvcEndpoint()
	if err != nil {
		return vtypes.MinerInfo{}, err
	}
	if env.Debug {
		pods, err := params.Venus.Pods(ctx)
		if err != nil {
			return vtypes.MinerInfo{}, err
		}

		svc, err := params.Venus.Svc(ctx)
		if err != nil {
			return vtypes.MinerInfo{}, err
		}

		endpoint, err = params.K8sEnv.PortForwardPod(ctx, pods[0].GetName(), int(svc.Spec.Ports[0].Port))
		if err != nil {
			return vtypes.MinerInfo{}, err
		}
	}
	client, closer, err := v1api.NewFullNodeRPC(ctx, endpoint.ToHttp(), nil)
	if err != nil {
		return vtypes.MinerInfo{}, err
	}
	defer closer()

	minfo, err := client.StateMinerInfo(ctx, maddr, vtypes.EmptyTSK)
	if err != nil {
		return vtypes.MinerInfo{}, err
	}
	return minfo, nil
}
