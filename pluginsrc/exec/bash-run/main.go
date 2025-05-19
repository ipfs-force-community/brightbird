package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"
)

var log = logging.Logger("bash")

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "bash",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "run bash script",
}

type TestCaseParams struct {
	InitParams *plugin.InitParams `jsonschema:"-" container:"type"`

	Script string `json:"script" jsonschema:"script"  title:"Script" require:"true" description:"bash script to run"`

	ParentDeployParams env.CommonDeployParams `json:"parentDeployParams" jsonschema:"parentDeployParams" title:"Parent Deploy Params" require:"false" description:"parent deploy params, used to get parent instance name"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	log.Infof("run script: \n------------------\n%v\n------------------\n", params.Script)
	log.Debugf("parentDeployParams: \n%v", params.ParentDeployParams)
	var execPodName string
	if len(params.ParentDeployParams.DeployName) > 0 {
		uid := env.UniqueId(k8sEnv.TestID(), k8sEnv.Retry(), params.ParentDeployParams.InstanceName)
		log.Debugf("pod search uid: %s", uid)
		if pods, err := k8sEnv.GetPodsByUid(ctx, uid); err == nil && len(pods) > 0 {
			for _, pod := range pods {
				log.Debug("pod : ", pod.GetName())
			}
			execPodName = pods[0].GetName()
		}
	}

	cmds := []string{"/bin/bash", "-c", params.Script}
	if params.InitParams != nil {
		initParamsBytes, err := json.Marshal(params.InitParams)
		if err != nil {
			return err
		}
		fmt.Println(string(initParamsBytes))
		cmds = append(cmds, "sh", string(initParamsBytes))
	}

	scriptBuf := bytes.NewBuffer(nil)
	stdOut := io.MultiWriter(os.Stdout, scriptBuf)

	if len(execPodName) > 0 {
		log.Infoln("exec in pod :", execPodName)
		if err := k8sEnv.ExecRemoteCmdWithStream(ctx, execPodName, true, stdOut, nil, cmds...); err != nil {
			return fmt.Errorf("pod %v exec failed : %w", execPodName, err)
		}
	} else {
		log.Infoln("exec in bash-run ")
		cmd := exec.Command(cmds[0], cmds[1:]...)
		cmd.Env = os.Environ()
		cmd.Stdout = stdOut
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	resultStr := plugin.GetLastJSON(scriptBuf.String())
	if len(resultStr) > 0 {
		log.Infoln("result : \n", resultStr)
		var result json.RawMessage
		if err := json.Unmarshal([]byte(resultStr), &result); err != nil {
			return fmt.Errorf("unmarshal bash result fail %w", err)
		}

		params.InitParams.Current().OutPut = result //set output
	}
	return nil
}
