package main

import (
	"context"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "add-miner",
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
	AdminToken types.AdminToken
	K8sEnv     *env.K8sEnvDeployer    `json:"-"`
	VenusAuth  env.IVenusAuthDeployer `json:"-"`
}

func Exec(ctx context.Context, params TestCaseParams) error {
	return nil
}
