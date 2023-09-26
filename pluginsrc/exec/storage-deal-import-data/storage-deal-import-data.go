package main

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	dropletmarket "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-market"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

var log = logging.Logger("storage-deal-import-data")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "storage-deal-import-data",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "手动导入deal数据",
}

type TestCaseParams struct {
	Droplet dropletmarket.DropletMarketDeployReturn `json:"Droplet" jsonschema:"Droplet" title:"Droplet" description:"droplet return"`

	ProposalCid string `json:"ProposalCid"  jsonschema:"ProposalCid"  title:"ProposalCid" require:"true" description:"ProposalCid"`
	CarFile     string `json:"carFile"  jsonschema:"carFile"  title:"carFile" require:"true" description:"carFile"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	proposalCid, err := cid.Decode(params.ProposalCid)
	if err != nil {
		return err
	}

	log.Debug("proposalCid: ", proposalCid)

	tmpl, err := template.New("command").Parse("./droplet storage deal import-data {{.ProposalCid}} {{.CarFile}}")
	if err != nil {
		return fmt.Errorf("parase template: %v", err)
	}

	data := map[string]interface{}{
		"ProposalCid": params.ProposalCid,
		"CarFile":     params.CarFile,
	}

	var importDataCmd bytes.Buffer
	err = tmpl.Execute(&importDataCmd, data)
	if err != nil {
		panic(err)
	}

	pods, err := dropletmarket.GetPods(ctx, k8sEnv, params.Droplet.InstanceName)
	if err != nil {
		return err
	}

	_, err = k8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), "/bin/sh", "-c", importDataCmd.String())
	if err != nil {
		return err
	}

	return nil
}
