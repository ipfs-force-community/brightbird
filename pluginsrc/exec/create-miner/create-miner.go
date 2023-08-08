package main

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/filecoin-project/go-address"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	damoclesmanager "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/damocles-manager"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("add-miner")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "create_miner",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "create miner address",
}

type TestCaseParams struct {
	Manager damoclesmanager.DamoclesManagerReturn `json:"DamoclesManager"  jsonschema:"DamoclesManager"  title:"Damocles Manager" require:"true" description:"manager return"`
	//todo support set owner/worker/controller
	WalletAddr address.Address `json:"walletAddr" jsonschema:"walletAddr" title:"Wallet Address" require:"true" description:"owner/worker address must be f3 address"`
}

type CreateMinerReturn struct {
	Miner  address.Address `json:"miner" jsonschema:"miner" title:"Miner Address" require:"true" description:"miner address"`
	Owner  address.Address `json:"owner" jsonschema:"owner" title:"Owner Address" require:"true" description:"owner address of miner"`
	Worker address.Address `json:"worker" jsonschema:"worker" title:"Worker Address" require:"true" description:"worker address of miner"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*CreateMinerReturn, error) {
	minerAddr, err := CreateMiner(ctx, k8sEnv, params, params.WalletAddr)
	if err != nil {
		fmt.Printf("create miner failed: %v\n", err)
		return nil, err
	}

	minerInfo, err := GetMinerInfo(ctx, k8sEnv, params, minerAddr)
	if err != nil {
		fmt.Printf("get miner info failed: %v\n", err)
		return nil, err
	}
	log.Debug("miner Info is %v", minerInfo)

	return &CreateMinerReturn{
		Miner:  minerAddr,
		Owner:  params.WalletAddr,
		Worker: params.WalletAddr,
	}, nil
}

func CreateMiner(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams, walletAddr address.Address) (address.Address, error) {
	damoclesPods, err := damoclesmanager.GetPods(ctx, k8sEnv, params.Manager.InstanceName)
	if err != nil {
		return address.Undef, err
	}
	cmd := []string{
		"./damocles-manager",
		"util",
		"miner",
		"create",
		"--sector-size=8MiB",
		"--exid=" + string(rune(rand.Intn(100000))),
	}
	cmd = append(cmd, "--from="+walletAddr.String())

	minerAddrStr, err := k8sEnv.ExecRemoteCmd(ctx, damoclesPods[0].GetName(), cmd...)
	if err != nil {
		return address.Undef, fmt.Errorf("exec remote cmd failed: %w", err)
	}

	return address.NewFromBytes(minerAddrStr)
}

func GetMinerInfo(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams, minerAddr address.Address) (string, error) {
	damoclesPods, err := damoclesmanager.GetPods(ctx, k8sEnv, params.Manager.InstanceName)
	if err != nil {
		return "", err
	}
	getMinerCmd := []string{
		"./damocles-manager",
		"util",
		"miner",
		"info",
		minerAddr.String(),
	}
	minerInfo, err := k8sEnv.ExecRemoteCmd(ctx, damoclesPods[0].GetName(), getMinerCmd...)
	if err != nil {
		return "", fmt.Errorf("exec remote cmd failed: %w", err)
	}

	return string(minerInfo), nil
}
