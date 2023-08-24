package plugin

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"regexp"

	container "github.com/golobby/container/v3"
	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/types"
	logging "github.com/ipfs/go-log/v2"
	"github.com/modern-go/reflect2"
	"github.com/tidwall/gjson"
)

var log = logging.Logger("plugin_setup")
var logDetail = &log.SugaredLogger

type InitParams struct {
	env.K8sInitParams

	env.EnvContext
}

func SetupPluginFromStdin(info types.PluginInfo, constructor interface{}) {
	cfg := logging.GetConfig()
	cfg.Stderr = false
	cfg.Stdout = true
	logging.SetupLogging(cfg) //set log to stdout

	fnT := reflect.TypeOf(constructor)
	depParmasT := fnT.In(2)
	inputSchema, err := ParserSchema(depParmasT)
	if err != nil {
		writeError(err)
		os.Exit(1)
		return
	}

	info.PluginParams = types.PluginParams{
		InputSchema: types.Schema(inputSchema),
	}
	if fnT.NumOut() == 2 {
		outputSchema, err := ParserSchema(fnT.Out(0))
		if err != nil {
			writeError(err)
			os.Exit(1)
			return
		}

		info.PluginParams.OutputSchema = types.Schema(outputSchema)
	}

	if len(os.Args) > 1 && os.Args[1] == "info" {
		result, err := json.Marshal(info)
		if err != nil {
			writeError(err)
			os.Exit(1)
			return
		}
		writeResult(string(result))
		return
	}

	reader := bufio.NewReader(os.Stdin)
	data, err := reader.ReadBytes('\n')
	if err != nil {
		writeError(err)
		os.Exit(1)
		return
	}

	incomingParams := &InitParams{}
	err = json.Unmarshal(data, incomingParams)
	if err != nil {
		writeError(err)
		os.Exit(1)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("%v", r)
			writeError(err)
			os.Exit(1)
		}
	}()

	err = runPlugin(info, constructor, incomingParams)
	if err != nil {
		writeError(err)
		os.Exit(1)
		return
	}
	os.Exit(0)
}

func runPlugin(info types.PluginInfo, constructor interface{}, incomingParams *InitParams) error {
	logLevel, err := incomingParams.Global.GetStringProperty("logLevel")
	if err != nil {
		return err
	}

	err = logging.SetLogLevel("*", logLevel)
	if err != nil {
		return err
	}

	nodeCtx := incomingParams.Current()
	input := nodeCtx.Input
	instanceName := gjson.GetBytes(input, "instanceName").Str
	//dump params
	data, err := json.Marshal(incomingParams)
	if err != nil {
		return err
	}
	logDetail = logDetail.With("plugin name", info.Name)
	logDetail.Infof("start running plugin params: %s", string(data))
	logDetail = logDetail.With("instance name", instanceName)

	k8sEnv, err := env.NewK8sEnvDeployer(incomingParams.K8sInitParams)
	if err != nil {
		return err
	}

	//set property
	fnT := reflect.TypeOf(constructor)
	depParmasT := fnT.In(2)
	paramsV := reflect.New(depParmasT)

	err = json.Unmarshal(input, paramsV.Interface())
	if err != nil {
		return err
	}

	err = container.Singleton(func() *env.GlobalParams {
		return &incomingParams.Global
	})
	if err != nil {
		return err
	}

	err = container.Singleton(func() *InitParams {
		return incomingParams
	})
	if err != nil {
		return err
	}

	err = container.Fill(paramsV.Interface())
	if err != nil {
		return err
	}

	//call function
	results := reflect.ValueOf(constructor).Call([]reflect.Value{reflect.ValueOf(context.Background()), reflect.ValueOf(k8sEnv), paramsV.Elem()})
	if len(results) == 1 {
		if !results[0].IsNil() {
			return results[0].Interface().(error)
		}
	}

	if len(results) == 2 {
		if !results[1].IsNil() {
			return results[1].Interface().(error)
		}

		if !reflect2.IsNil(results[0].Interface()) {
			jsonBytes, err := json.Marshal(results[0].Interface())
			if err != nil {
				return err
			}

			nodeCtx.OutPut = jsonBytes //override output
		}
	}

	jsonBytes, err := json.Marshal(incomingParams)
	if err != nil {
		return err
	}
	writeResult(string(jsonBytes)) //must return this value
	return nil
}

func GetPluginInfo(path string) (*types.PluginInfo, error) {
	stdOut := bytes.NewBuffer(nil)

	cmd := exec.Command(path, "info")
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = io.MultiWriter(os.Stdout, stdOut)
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	info := &types.PluginInfo{}
	err = json.Unmarshal(stdOut.Bytes(), info)
	if err != nil {
		return nil, err
	}
	reg, err := regexp.Compile(`v[\d]+?.[\d]+?.[\d]+?`)
	if err != nil {
		return nil, err
	}
	version := reg.FindString(info.Version)
	if len(version) == 0 {
		return nil, fmt.Errorf("not validate version string %s. must contain substring vx.x.x", info.Version)
	}
	info.Version = version
	return info, nil
}

func writeError(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
}

func writeResult(val string) {
	fmt.Fprintln(os.Stdout, val)
}
