package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "add_miner",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "user add miner",
}

type TestCaseParams struct {
	fx.In
	Params struct {
		From       string `json:"from"`
		Owner      string `json:"owner"`
		Worker     string `json:"worker"`
		SectorSize string `json:"sectorSize"`
		Peer       string `json:"peer"`
		Multiaddr  string `json:"multiaddr"`
		Exid       string `json:"exid"`
	} `optional:"true"`
	AdminToken                 types.AdminToken
	K8sEnv                     *env.K8sEnvDeployer             `json:"-"`
	VenusSectorManagerDeployer env.IVenusSectorManagerDeployer `json:"-"`
}

func Exec(ctx context.Context, params TestCaseParams) error {
	cmd := []string{
		"./venus-sector-manager",
		"util",
		"miner",
		"create",
	}
	switch {
	case params.Params.From != "":
		cmd = append(cmd, "--from="+params.Params.From)
	case params.Params.Owner != "":
		cmd = append(cmd, "--owner="+params.Params.Owner)
	case params.Params.Worker != "":
		cmd = append(cmd, "--worker="+params.Params.Worker)
	case params.Params.SectorSize != "":
		cmd = append(cmd, "--peer="+params.Params.Peer)
	case params.Params.Multiaddr != "":
		cmd = append(cmd, "--multiaddr="+params.Params.Multiaddr)
	case params.Params.Exid != "":
		cmd = append(cmd, "--exid"+params.Params.Exid)
	default:
		return errors.New("parameter err")
	}

	fmt.Println(cmd)
	minerAddr, err := params.K8sEnv.ExecRemoteCmd(ctx, params.VenusSectorManagerDeployer.Pods()[0].GetName(), cmd)
	if err != nil {
		return err
	}

	getMinerCmd := []string{
		"./venus-sector-manager",
		"util",
		"miner",
		"info",
		string(minerAddr),
	}
	minerInfo, err := params.K8sEnv.ExecRemoteCmd(ctx, params.VenusSectorManagerDeployer.Pods()[0].GetName(), getMinerCmd)
	if err != nil {
		return err
	}
	fmt.Println(minerInfo)
	return nil
}
