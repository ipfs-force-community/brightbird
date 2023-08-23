package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
	logging "github.com/ipfs/go-log/v2"
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
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	log.Info("run script \n", params.Script) // xxxx\n
	initParamsBytes, err := json.Marshal(params.InitParams)
	if err != nil {
		return err
	}

	fmt.Println(string(initParamsBytes))
	cmd := exec.Command("/bin/bash", "-c", params.Script, "sh", string(initParamsBytes))
	cmd.Env = os.Environ()

	scriptBuf := bytes.NewBuffer(nil)
	cmd.Stdout = io.MultiWriter(os.Stdout, scriptBuf)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	resultStr := plugin.GetLastJSON(scriptBuf.String())
	fmt.Println("result ", scriptBuf.String())
	fmt.Println("result ", resultStr)
	if len(resultStr) > 0 {
		var result json.RawMessage
		err = json.Unmarshal([]byte(resultStr), &result)
		if err != nil {
			return fmt.Errorf("unmarshal bash result fail %w", err)
		}

		params.InitParams.Current().OutPut = result //set output

	}
	return nil
}
