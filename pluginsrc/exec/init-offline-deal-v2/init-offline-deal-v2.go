package main

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/venus-shared/actors/policy"
	clientapi "github.com/filecoin-project/venus/venus-shared/api/market/client"
	"github.com/filecoin-project/venus/venus-shared/types/market/client"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"

	dropletclient "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/droplet-client"
)

var log = logging.Logger("init-offline-deal-v2")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "init-offline-deal-v2",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "droplet-client init-v2 offline deal",
}

type TestCaseParams struct {
	DropletClient dropletclient.DropletClientDeployReturn `json:"DropClient" jsonschema:"DropClient" title:"DropletClient" description:"droplet client return"`

	File    string `json:"File" jsonschema:"File" title:"File"`
	CarFile string `json:"CarFile" jsonschema:"CarFile" title:"CarFile"`

	MinerAddress address.Address `json:"minerAddress"  jsonschema:"minerAddress" title:"MinerAddress" require:"true"`
	Duration     int64           `json:"Duration"  jsonschema:"Duration"  title:"Duration" default:"518400" require:"true" description:"Set the price of the ask for retrievals"`
	From         address.Address `json:"From"  jsonschema:"From"  title:"From" require:"false" description:"From"`
}

type InitOfflineDealReturn struct {
	ProposalCid string
	DealUUID    string
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) (*InitOfflineDealReturn, error) {
	api, closer, err := clientapi.DialIMarketClientRPC(ctx, params.DropletClient.SvcEndpoint.ToMultiAddr(), params.DropletClient.ClientToken, nil)
	if err != nil {
		return nil, fmt.Errorf("dial market client rpc failed")
	}
	defer closer()

	initData, err := getCidAndPieceSize(ctx, api, params.File, params.CarFile)
	if err != nil {
		return nil, err
	}

	log.With(
		"Root", initData.Root,
		"Cid", initData.Cid,
		"PieceSize", initData.PieceSize,
	).Info("piece info")

	wallet := params.From
	if wallet == address.Undef {
		wallet, err = api.DefaultAddress(ctx)
		if err != nil {
			return nil, err
		}
	}

	duration := params.Duration
	MinDealDuration, MaxDealDuration := policy.DealDurationBounds(0)
	if abi.ChainEpoch(duration) < MinDealDuration {
		return nil, fmt.Errorf("minimum deal duration is %d blocks", MinDealDuration)
	}
	if abi.ChainEpoch(duration) > MaxDealDuration {
		return nil, fmt.Errorf("maximum deal duration is %d blocks", MaxDealDuration)
	}

	tmpl, err := template.New("command").Parse(`./droplet-client storage deals init-v2 \
	--provider={{.Miner}} \
	--from={{.From}} \
	--piece-cid={{.PieceCid}} \
	--piece-size={{.PieceSize}} \
	--payload-cid={{.DataCid}} \
	--duration={{.Duration}} \
	--storage-price=0`)
	if err != nil {
		return nil, fmt.Errorf("parse template: %v", err)
	}

	data := map[string]interface{}{
		"From":      wallet,
		"PieceCid":  initData.Cid,
		"PieceSize": initData.PieceSize,
		"DataCid":   initData.Root,
		"Miner":     params.MinerAddress,
		"Duration":  duration,
	}

	var initDealCmd bytes.Buffer
	if err := tmpl.Execute(&initDealCmd, data); err != nil {
		return nil, fmt.Errorf("execute template: %v", err)
	}

	pods, err := dropletclient.GetPods(ctx, k8sEnv, params.DropletClient.InstanceName)
	if err != nil {
		return nil, err
	}
	if len(pods) == 0 {
		return nil, fmt.Errorf("not found pod for %v", params.DropletClient.InstanceName)
	}
	execPodName := pods[0].GetName()
	res, err := k8sEnv.ExecRemoteCmd(ctx, execPodName, "/bin/sh", "-c", initDealCmd.String())
	if err != nil {
		return nil, fmt.Errorf("exec [%v] failed: %v", initDealCmd.String(), err)
	}

	log.Infof("init deal result: %s", res)

	var dealUUID, proposalCid string
	for _, line := range strings.Split(string(res), "\n") {
		if strings.Contains(line, "deal") && strings.Contains(line, "uuid:") {
			dealUUID = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(line, "deal", ""), "uuid:", ""))
		}
		if strings.Contains(line, "proposal") && strings.Contains(line, "cid:") {
			proposalCid = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(line, "proposal", ""), "cid:", ""))
		}
	}

	if len(proposalCid) == 0 {
		dealStatusCmd := fmt.Sprintf("./droplet-client storage deals status  --provider %v --deal-uuid %v", params.MinerAddress.String(), dealUUID)
		res, err := k8sEnv.ExecRemoteCmd(ctx, execPodName, "/bin/sh", "-c", dealStatusCmd)
		if err != nil {
			return nil, fmt.Errorf("exec [%v] failed: %v", initDealCmd, err)
		}
		for _, line := range strings.Split(string(res), "\n") {
			if strings.Contains(line, "proposal") && strings.Contains(line, "cid:") {
				proposalCid = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(line, "proposal", ""), "cid:", ""))
				break
			}
		}
	}
	log.With(
		"uuid", dealUUID,
		"proposal", proposalCid,
	).Info("res")

	return &InitOfflineDealReturn{
		ProposalCid: proposalCid,
		DealUUID:    dealUUID,
	}, nil
}

type DataImportReturn struct {
	Root      cid.Cid
	Cid       cid.Cid
	PieceSize abi.UnpaddedPieceSize
}

func getCidAndPieceSize(ctx context.Context, api clientapi.IMarketClient, file, carFile string) (*DataImportReturn, error) {
	absPath, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}

	ref := client.FileRef{
		Path:  absPath,
		IsCAR: false,
	}
	c, err := api.ClientImport(ctx, ref)
	if err != nil {
		return nil, err
	}

	ref2 := client.FileRef{
		Path:  file,
		IsCAR: false,
	}

	err = api.ClientGenCar(ctx, ref2, carFile)
	if err != nil {
		return nil, err
	}

	ret, err := api.ClientCalcCommP(ctx, carFile)
	if err != nil {
		return nil, err
	}
	return &DataImportReturn{
		Root:      c.Root,
		Cid:       ret.Root,
		PieceSize: ret.Size,
	}, nil
}
