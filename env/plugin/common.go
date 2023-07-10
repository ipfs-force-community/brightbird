package plugin

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	logging "github.com/ipfs/go-log/v2"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var log = logging.Logger("plugin_setup")
var logDetail = &log.SugaredLogger

type InitParams struct {
	env.K8sInitParams

	env.EnvContext

	//current
	//CodeVersion    string          //todo allow config as tag commit id brance
	//InstanceName   string          //plugin instance name
	//PropertiesJson json.RawMessage // get params form this fields
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
		writeResult(string(result)) //NOTE never remove this print code!!!!!, println for testrunner to read
		return
	}

	reader := bufio.NewReader(os.Stdin)
	data, err := reader.ReadBytes('\n')
	if err != nil {
		writeError(err)
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
	err := logging.SetLogLevel("*", incomingParams.Global.LogLevel)
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

	sjson.SetBytes(input, "global", incomingParams.Global)
	err = json.Unmarshal(input, paramsV.Interface())
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

		jsonBytes, err := json.Marshal(results[0].Interface())
		if err != nil {
			return err
		}

		nodeCtx.OutPut = jsonBytes
	}

	jsonBytes, err := json.Marshal(incomingParams)
	if err != nil {
		return err
	}
	writeResult(string(jsonBytes))
	return nil
}

func GetPluginInfo(path string) (*types.PluginInfo, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	process, err := os.StartProcess(path, []string{path, "info"}, &os.ProcAttr{
		Files: []*os.File{os.Stdin, w, os.Stderr},
	})
	if err != nil {
		return nil, err
	}

	st, err := process.Wait()
	if err != nil {
		return nil, err
	}

	if st.ExitCode() != 0 {
		return nil, fmt.Errorf("get detail of plugin %s fail exitcode %d", path, st.ExitCode())
	}

	w.Close() //nolint
	bufR := bufio.NewReader(io.TeeReader(r, os.Stdout))
	var lastLine string
	for {
		thisLine, err := bufR.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		lastLine = thisLine
	}

	info := &types.PluginInfo{}
	err = json.Unmarshal([]byte(lastLine), info)
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
