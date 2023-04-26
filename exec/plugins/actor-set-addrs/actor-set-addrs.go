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
	"github.com/filecoin-project/venus/venus-shared/api/wallet"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	vtypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"go.uber.org/fx"
	"math/rand"
	"strings"
)

var Info = types.PluginInfo{
	Name:        "actor-set-addrs",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "actor set-addrs",
}

type TestCaseParams struct {
	fx.In
	AdminToken                 types.AdminToken
	K8sEnv                     *env.K8sEnvDeployer             `json:"-"`
	VenusAuth                  env.IVenusAuthDeployer          `json:"-"`
	VenusMarket                env.IVenusMarketDeployer        `json:"-"`
	VenusWallet                env.IVenusWalletDeployer        `json:"-" svcname:"Wallet"`
	VenusMiner                 env.IVenusMinerDeployer         `json:"-"`
	VenusSectorManagerDeployer env.IVenusSectorManagerDeployer `json:"-"`
	Venus                      env.IVenusDeployer              `json:"-"`
}

func Exec(ctx context.Context, params TestCaseParams) error {

	listenAddress, err := marketListen(ctx, params)
	if err != nil {
		fmt.Printf("market net listen err: %v\n", err)
		return err
	}
	fmt.Printf("market net listen is: %v\n", listenAddress)

	walletAddr, err := CreateWallet(ctx, params)
	if err != nil {
		fmt.Printf("create wallet failed: %v\n", err)
		return err
	}

	minerAddr, err := CreateMiner(ctx, params, walletAddr)
	if err != nil {
		fmt.Printf("create miner failed: %v\n", err)
		return err
	}

	messageId, err := SetActorAddr(ctx, params, minerAddr)
	if err != nil {
		fmt.Printf("set actor address failed: %v\n", err)
		return err
	}
	fmt.Printf("set actor address message id is: %v\n", messageId)

	return nil
}

func marketListen(ctx context.Context, params TestCaseParams) (string, error) {
	endpoint := params.VenusMarket.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, params.VenusMarket.Pods()[0].GetName(), int(params.VenusMarket.Svc().Spec.Ports[0].Port))
		if err != nil {
			return "", err
		}
	}
	client, closer, err := marketapi.NewIMarketRPC(ctx, endpoint.ToHttp(), nil)
	if err != nil {
		return "", err
	}
	defer closer()

	addrs, err := client.NetAddrsListen(ctx)
	if err != nil {
		return "", nil
	}

	for _, peer := range addrs.Addrs {
		fmt.Printf("%s/p2p/%s\n", peer, addrs.ID)
	}
	return "", err
}

func CreateWallet(ctx context.Context, params TestCaseParams) (address.Address, error) {
	walletToken, err := env.ReadWalletToken(ctx, params.K8sEnv, params.VenusWallet.Pods()[0].GetName())
	if err != nil {
		return address.Undef, fmt.Errorf("read wallet token failed: %w\n", err)
	}

	endpoint := params.VenusWallet.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, params.VenusWallet.Pods()[0].GetName(), int(params.VenusWallet.Svc().Spec.Ports[0].Port))
		if err != nil {
			return address.Undef, fmt.Errorf("port forward failed: %w\n", err)
		}
	}

	walletRpc, closer, err := wallet.DialIFullAPIRPC(ctx, endpoint.ToMultiAddr(), walletToken, nil)
	if err != nil {
		return address.Undef, fmt.Errorf("dial iFullAPI rpc failed: %w\n", err)
	}
	defer closer()

	password := "123456"
	err = walletRpc.SetPassword(ctx, password)
	if err != nil {
		return address.Undef, fmt.Errorf("set password failed: %w\n", err)
	}

	walletAddr, err := walletRpc.WalletNew(ctx, vTypes.KTBLS)
	if err != nil {
		return address.Undef, fmt.Errorf("create wallet failed: %w\n", err)
	}
	fmt.Printf("wallet: %v\n", walletAddr)

	return walletAddr, nil
}

func CreateMiner(ctx context.Context, params TestCaseParams, walletAddr address.Address) (string, error) {
	cmd := []string{
		"./venus-sector-manager",
		"util",
		"miner",
		"create",
		"--sector-size=8MiB",
		"--exid=" + string(rune(rand.Intn(1000000))),
	}
	cmd = append(cmd, "--from="+walletAddr.String())

	minerAddr, err := params.K8sEnv.ExecRemoteCmd(ctx, params.VenusSectorManagerDeployer.Pods()[0].GetName(), cmd)
	if err != nil {
		return "", fmt.Errorf("exec remote cmd failed: %w\n", err)
	}

	return string(minerAddr), nil
}

func SetActorAddr(ctx context.Context, params TestCaseParams, minerAddr string) (string, error) {
	endpoint := params.VenusMarket.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, params.VenusMarket.Pods()[0].GetName(), int(params.VenusMarket.Svc().Spec.Ports[0].Port))
		if err != nil {
			return "", err
		}
	}
	client, closer, err := marketapi.NewIMarketRPC(ctx, endpoint.ToHttp(), nil)
	if err != nil {
		return "", err
	}
	defer closer()

	addrs, err := client.NetAddrsListen(ctx)
	if err != nil && addrs.Addrs != nil {
		return addrs.String(), nil
	}

	MessageParams, err := ConstructParams(addrs)
	if err != nil {
		return "", err
	}

	maddr, err := address.NewFromString(minerAddr)
	if err != nil {
		return "", nil
	}

	minfo, err := GetMinerInfo(ctx, params, maddr)
	if err != nil {
		return "", err
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
		return "", err
	}

	fmt.Printf("Requested multiaddrs change in message %s\n", mid)

	return "", err
}

func ConstructParams(address peer.AddrInfo) (param []byte, err error) {
	addr := ""
	for _, peer := range address.Addrs {
		if strings.HasPrefix(peer.String(), "/ip4/192") {
			addr = peer.String()
			break
		}
	}

	maddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %q as a multiaddr: %w", addr, err)
	}
	maddrNop2p, strip := ma.SplitFunc(maddr, func(c ma.Component) bool {
		return c.Protocol().Code == ma.P_P2P
	})
	if strip != nil {
		fmt.Println("Stripping peerid ", strip, " from ", maddr)
	}

	var addrs []abi.Multiaddrs
	addrs = append(addrs, maddrNop2p.Bytes())

	params, err := actors.SerializeParams(&vtypes.ChangeMultiaddrsParams{NewMultiaddrs: addrs})
	if err != nil {
		return nil, err
	}
	return params, nil
}

func GetMinerInfo(ctx context.Context, params TestCaseParams, maddr address.Address) (vtypes.MinerInfo, error) {
	endpoint := params.Venus.SvcEndpoint()
	if env.Debug {
		var err error
		endpoint, err = params.K8sEnv.PortForwardPod(ctx, params.Venus.Pods()[0].GetName(), int(params.Venus.Svc().Spec.Ports[0].Port))
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
