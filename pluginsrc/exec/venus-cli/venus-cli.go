package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"

	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "venus-cli",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "venus cli test",
}

type TestCaseParams struct {
	Venus  venus.VenusDeployReturn `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	Params []string                `json:"params" jsonschema:"params" title:"Params" require:"true" description:"venus cli params: /app/venus [xxxx]"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	pods, err := venus.GetPods(ctx, k8sEnv, params.Venus.InstanceName)
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return fmt.Errorf("no pod found for %s", params.Venus.InstanceName)
	}
	cmdStr := "/app/venus"
	for _, par := range params.Params {
		cmdStr += " " + par
	}
	fmt.Println("exec cmd:", cmdStr)

	for i := 0; i < 10; i++ {
		if _, err := k8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), "/bin/bash", "-c", cmdStr); err == nil {
			break
		}
		fmt.Println("exec cmd failed, retrying... ", i)
		time.Sleep(time.Second)
	}
	res, err := k8sEnv.ExecRemoteCmd(ctx, pods[0].GetName(), "/bin/bash", "-c", cmdStr)
	if err != nil {
		return fmt.Errorf("exec cmd failed: %w", err)
	}
	fmt.Println(string(res))
	return nil
}
