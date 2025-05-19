package main

import (
	"context"
	"fmt"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"

	damoclesworker "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/damocles-worker"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "damocles-worker-cli",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "damocles-worker cli 相关测试",
}

type TestCaseParams struct {
	DamoclesWorker damoclesworker.DamoclesWorkerReturn `json:"DamoclesWorker" jsonschema:"DamoclesWorker" title:"Damocles worker daemon" require:"true" description:"damocles-worker deploy return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	pods, err := damoclesworker.GetPods(ctx, k8sEnv, params.DamoclesWorker.InstanceName)
	if err != nil {
		return err
	}

	workerListResult, err := k8sEnv.ExecRemoteCmd(ctx, pods[0].Name, "/damocles-worker", "worker", "list")
	if err != nil {
		return fmt.Errorf("exec `worker list`: %w", err)
	}

	fmt.Println(string(workerListResult))
	return nil
}
