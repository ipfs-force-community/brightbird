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

	data := []byte(`{"Namespace":"li","TestID":"95b41d9e","PrivateRegistry":"192.168.200.175","MysqlConnTemplate":"root:Aa123456@(192.168.200.175:3306)/%s?parseTime=true\u0026loc=Local\u0026charset=utf8mb4\u0026collation=utf8mb4_unicode_ci\u0026readTimeout=10s\u0026writeTimeout=10s","TmpPath":"/shared-dir/tmp","Global":{"logLevel":"DEBUG","customProperties":{"BootstrapPeer":["/ip4/192.168.200.125/tcp/34567/p2p/12D3KooWB1X6MKuvZGN15YMJFMDSFLDeSZyCEiiuRV6Wyucq3bAZ"]}},"Nodes":{"create_token-6e488a84":{"Input":{"SophonAuth":{"adminToken":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiYWRtaW4iLCJwZXJtIjoiYWRtaW4iLCJleHQiOiIifQ.bOA7-uzkL_WF5S6TMhznaODTzJQ1bKkF-U71SM_sZXA","codeVersion":"f70d19c50f0005949b7f239e3a5cae804fae5496","configMapName":"sophon-auth-95b41d9ef1116e8a","deployName":"sophon-auth","instanceName":"sophon-auth-96e7cb7d","mysqlDSN":"root:Aa123456@(192.168.200.175:3306)/sophon-auth-95b41d9ef1116e8a?parseTime=true\u0026loc=Local\u0026charset=utf8mb4\u0026collation=utf8mb4_unicode_ci\u0026readTimeout=10s\u0026writeTimeout=10s","replicas":1,"statefulSetName":"sophon-auth-95b41d9ef1116e8a","svcEndpoint":"sophon-auth-95b41d9ef1116e8a-service:8989","svcName":"sophon-auth-95b41d9ef1116e8a-service"},"codeVersion":"","extra":"nono","instanceName":"create_token-6e488a84","perm":"read","userName":"li"},"OutPut":{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoibGkiLCJwZXJtIjoicmVhZCIsImV4dCI6Im5vbm8ifQ.7xLm_13pFVspa226ZVQW4TK2heSIaXQ9c7bU66Fr9eQ"}},"create_user-ecc4dbce":{"Input":{"SophonAuth":{"adminToken":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiYWRtaW4iLCJwZXJtIjoiYWRtaW4iLCJleHQiOiIifQ.bOA7-uzkL_WF5S6TMhznaODTzJQ1bKkF-U71SM_sZXA","codeVersion":"f70d19c50f0005949b7f239e3a5cae804fae5496","configMapName":"sophon-auth-95b41d9ef1116e8a","deployName":"sophon-auth","instanceName":"sophon-auth-96e7cb7d","mysqlDSN":"root:Aa123456@(192.168.200.175:3306)/sophon-auth-95b41d9ef1116e8a?parseTime=true\u0026loc=Local\u0026charset=utf8mb4\u0026collation=utf8mb4_unicode_ci\u0026readTimeout=10s\u0026writeTimeout=10s","replicas":1,"statefulSetName":"sophon-auth-95b41d9ef1116e8a","svcEndpoint":"sophon-auth-95b41d9ef1116e8a-service:8989","svcName":"sophon-auth-95b41d9ef1116e8a-service"},"codeVersion":"","comment":"hei","instanceName":"create_user-ecc4dbce","userName":"li"},"OutPut":{"userName":"li"}},"sophon-auth-96e7cb7d":{"Input":{"codeVersion":"f70d19c50f0005949b7f239e3a5cae804fae5496","instanceName":"sophon-auth-96e7cb7d","replicas":1},"OutPut":{"mysqlDSN":"root:Aa123456@(192.168.200.175:3306)/sophon-auth-95b41d9ef1116e8a?parseTime=true\u0026loc=Local\u0026charset=utf8mb4\u0026collation=utf8mb4_unicode_ci\u0026readTimeout=10s\u0026writeTimeout=10s","replicas":1,"adminToken":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiYWRtaW4iLCJwZXJtIjoiYWRtaW4iLCJleHQiOiIifQ.bOA7-uzkL_WF5S6TMhznaODTzJQ1bKkF-U71SM_sZXA","codeVersion":"f70d19c50f0005949b7f239e3a5cae804fae5496","instanceName":"sophon-auth-96e7cb7d","deployName":"sophon-auth","statefulSetName":"sophon-auth-95b41d9ef1116e8a","configMapName":"sophon-auth-95b41d9ef1116e8a","svcName":"sophon-auth-95b41d9ef1116e8a-service","svcEndpoint":"sophon-auth-95b41d9ef1116e8a-service:8989"}},"venus-daemon-b7a8fad0":{"Input":{"SophonAuth":{"adminToken":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiYWRtaW4iLCJwZXJtIjoiYWRtaW4iLCJleHQiOiIifQ.bOA7-uzkL_WF5S6TMhznaODTzJQ1bKkF-U71SM_sZXA","codeVersion":"f70d19c50f0005949b7f239e3a5cae804fae5496","configMapName":"sophon-auth-95b41d9ef1116e8a","deployName":"sophon-auth","instanceName":"sophon-auth-96e7cb7d","mysqlDSN":"root:Aa123456@(192.168.200.175:3306)/sophon-auth-95b41d9ef1116e8a?parseTime=true\u0026loc=Local\u0026charset=utf8mb4\u0026collation=utf8mb4_unicode_ci\u0026readTimeout=10s\u0026writeTimeout=10s","replicas":1,"statefulSetName":"sophon-auth-95b41d9ef1116e8a","svcEndpoint":"sophon-auth-95b41d9ef1116e8a-service:8989","svcName":"sophon-auth-95b41d9ef1116e8a-service"},"codeVersion":"d4b62f69831068f69e606544e7d489e2c67e14cb","instanceName":"venus-daemon-b7a8fad0","netType":"force","replicas":1},"OutPut":{}}},"CurrentContext":"venus-daemon-b7a8fad0"}	`)

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

	input, err = sjson.SetBytes(input, "global", incomingParams.Global)
	if err != nil {
		return err
	}

	err = json.Unmarshal(input, paramsV.Interface())
	if err != nil {
		return err
	}
	xx := paramsV.Interface()
	fmt.Println(xx)

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
