package plugin

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/fx_opt"
	"github.com/hunjixin/brightbird/types"
	logging "github.com/ipfs/go-log/v2"
	"github.com/modern-go/reflect2"
	"go.uber.org/fx"
)

var log = logging.Logger("plugin_setup")
var logDetail = &log.SugaredLogger

type InitParams struct {
	env.K8sInitParams
	BoostrapPeers types.BootstrapPeers
	CodeVersion   string                      //todo allow config as tag commit id brance
	InstanceName  string                      //plugin instance name
	SockPath      string                      //path to listen
	Dependencies  []*types.DependencyProperty // get dependencies
	Properties    []*types.Property           // get params form this fields
}

func SetupPluginFromStdin(info types.PluginInfo, constructor interface{}) {
	err := logging.SetLogLevel("*", "DEBUG")
	if err != nil {
		fmt.Println("init log fail", err.Error())
		os.Exit(1)
		return
	}

	logDetail = logDetail.With("plugin name", info.Name)
	logDetail.Infof("start running plugin")
	fnT := reflect.TypeOf(constructor)
	depParmasT := fnT.In(1)
	pluginParams, err := ParseParams(depParmasT)
	if err != nil {
		respError(err)
		os.Exit(1)
		return
	}
	info.PluginParams = pluginParams

	if len(os.Args) > 1 && os.Args[1] == "info" {
		respJSON(info) //NOTE never remove this print code!!!!!, println for testrunner to read
		return
	}

	reader := bufio.NewReader(os.Stdin)
	data, err := reader.ReadBytes('\n')

	if err != nil {
		respError(err)
		return
	}

	//data = []byte(`{"Namespace":"production","TestID":"e35e2b9f","PrivateRegistry":"192.168.200.175","MysqlConnTemplate":"root:Aa123456@(192.168.200.175:3306)/%s?parseTime=true\u0026loc=Local\u0026charset=utf8mb4\u0026collation=utf8mb4_unicode_ci\u0026readTimeout=10s\u0026writeTimeout=10s","TmpPath":"","codeVersion":"","InstanceName":"set-pwd","SockPath":"/tmp/e35e2b9f_set-pwd_e6dd33fd-24cb-4376-a21e-9fd3e66ee425.sock799133915","Dependencies":[{"name":"VenusWallet","value":"wallet","type":"Deployer","sockPath":"/tmp/e35e2b9f_wallet_22c7dcbf-55ca-4edf-a6e3-9e6832cbce1c.sock","description":"","require":false}],"Properties":[{"name":"password","type":"string","description":"","value":"123","require":false}]}`)
	incomingParams := &InitParams{}
	err = json.Unmarshal(data, incomingParams)
	if err != nil {
		respError(err)
		os.Exit(1)
		return
	}

	logDetail = logDetail.With("instance name", incomingParams.InstanceName)

	callFn, err := ConvertDeployConstructor(incomingParams, info, reflect.ValueOf(constructor), depParmasT)
	if err != nil {
		respError(err)
		os.Exit(1)
		return
	}

	ctx := context.Background()
	_, err = fx_opt.New(ctx,
		fx_opt.Override(new(context.Context), ctx),
		fx_opt.Override(new(*InitParams), incomingParams),
		fx_opt.Override(new(types.BootstrapPeers), incomingParams.BoostrapPeers),
		fx_opt.Override(new(*env.K8sEnvDeployer), func() (*env.K8sEnvDeployer, error) {
			return env.NewK8sEnvDeployer(incomingParams.K8sInitParams)
		}),
		func() fx_opt.Option {
			var opts []fx_opt.Option
			//todo make simplify
			for _, dependency := range incomingParams.Dependencies {
				if !dependency.Require && len(dependency.Value) == 0 {
					continue
				}

				switch dependency.Type {
				case types.Deploy:
					fn := MakeDepoloyDepedency(dependency)
					opts = append(opts, fx_opt.Annotate(struct {
						ty   reflect.Type
						name string
					}{
						ty:   env.IDeployerT,
						name: dependency.Value,
					}, fn))
				case types.TestExec:
					fn := MakeExecDepedency(dependency)
					opts = append(opts, fx_opt.Annotate(struct {
						ty   reflect.Type
						name string
					}{
						ty:   env.IExecT,
						name: dependency.Value,
					}, fn))
				default:
					opts = append(opts, fx_opt.Error(errors.New("unsupport plugin type")))
				}
			}
			return fx_opt.Options(opts...)
		}(),
		fx_opt.If(info.PluginType == types.Deploy,
			fx_opt.Override(new(env.IDeployer), callFn),
			fx_opt.Override(fx_opt.NextInvoke(), StartDeployPlugin),
		),

		fx_opt.If(info.PluginType == types.TestExec,
			fx_opt.Override(new(env.IExec), callFn),
			fx_opt.Override(fx_opt.NextInvoke(), StartExecPlugin),
		),
	)
	if err != nil {
		logDetail.Infof("run plugin fail")
		respError(err)
		os.Exit(1)
		return
	}
	logDetail.Infof("run plugin successfully")
	os.Exit(0)
}

func GetPluginInfo(path string) (*types.PluginInfo, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer w.Close() //nolint

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

	reader := bufio.NewReader(io.TeeReader(r, os.Stdout))
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}
		cmd, val, isCmd := ReadCMD(string(line))
		if isCmd {
			if cmd == CMDVALPREFIX {
				info := &types.PluginInfo{}
				err = json.Unmarshal([]byte(val), info)
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
		}

		if err == io.EOF {
			break
		}
	}
	return nil, errors.New("not found info value")
}

func ConvertDeployConstructor(incomingParams *InitParams, pluginInfo types.PluginInfo, ctorFn reflect.Value, depParmasT reflect.Type) (interface{}, error) {
	svcMap, err := getSvcMap(incomingParams.Dependencies...)
	if err != nil {
		return nil, err
	}

	fnT := ctorFn.Type()
	newInStruct := convertInjectParams(depParmasT, svcMap)

	var newOutArgs []reflect.Type
	if pluginInfo.PluginType == types.Deploy {
		newOutArgs = []reflect.Type{env.IDeployerT, types.ErrT}
	} else {
		newOutArgs = []reflect.Type{env.IExecT, types.ErrT}
	}

	newFn := reflect.FuncOf([]reflect.Type{types.CtxT, newInStruct}, newOutArgs, false)
	return reflect.MakeFunc(newFn, func(args []reflect.Value) (vals []reflect.Value) {
		defer func() {
			if r := recover(); r != nil {
				vals = make([]reflect.Value, 2)
				vals[0] = reflect.Zero(newOutArgs[0])
				logDetail.Info("stacktrace from panic:" + string(debug.Stack()))
				vals[1] = reflect.ValueOf(fmt.Errorf("invoke plugin %s %v", pluginInfo.Name, r))
			} else {
				logDetail.Infof("completed plugin %s name: %s", pluginInfo.Name, incomingParams.InstanceName)
			}
		}()

		logDetail.Infof("start to deploy %s name: %s", pluginInfo.Name, incomingParams.InstanceName)
		//convert params
		argT := fnT.In(1)
		dstVal := reflect.New(argT).Elem()
		for j := 0; j < argT.NumField(); j++ {
			field := argT.Field(j)
			fieldName := field.Name
			if !field.Anonymous && len(fieldName) != 0 {
				if fieldName == "Params" {
					val := reflect.New(field.Type)
					err := collectParams(incomingParams.Properties, val.Interface())
					if err != nil {
						return []reflect.Value{reflect.Zero(newOutArgs[0]), reflect.ValueOf(err)}
					}

					baseConfig := val.Elem().FieldByName("BaseConfig")
					if (baseConfig != reflect.Value{}) {
						baseConfig.Set(reflect.ValueOf(env.NewBaseConfig(incomingParams.CodeVersion, incomingParams.InstanceName)))
					}

					dstVal.FieldByName(fieldName).Set(val.Elem())
				} else {
					dstVal.FieldByName(fieldName).Set(args[1].FieldByName(fieldName))
				}
			}
		}

		//call plugin
		results := ctorFn.Call([]reflect.Value{args[0], dstVal})
		//convert result
		if !results[1].IsNil() {
			return []reflect.Value{reflect.Zero(newOutArgs[0]), results[1]}
		}
		destResultVal := reflect.New(newOutArgs[0]).Elem()
		destResultVal.Set(results[0])
		return []reflect.Value{destResultVal, results[1]}
	}).Interface(), nil
}

func MakeDepoloyDepedency(property *types.DependencyProperty) interface{} {
	fileds := []reflect.StructField{
		{
			Name:      "Out",
			Type:      reflect.TypeOf(fx.Out{}),
			Offset:    1,
			Index:     []int{int(1)},
			Anonymous: true,
		},
	}

	if len(property.Value) > 0 {
		tag := ""
		if !property.Require {
			tag = `optional:"true"`
		}
		tag = tag + fmt.Sprintf(` name:"%s"`, property.Value)
		fileds = append(fileds, reflect.StructField{
			Name:      "OutVal",
			Type:      env.IDeployerT,
			Tag:       reflect.StructTag(tag),
			Offset:    0,
			Index:     []int{int(0)},
			Anonymous: false,
		})
	}

	outType := reflect.StructOf(fileds)
	fnT := reflect.FuncOf([]reflect.Type{}, []reflect.Type{outType, types.ErrT}, false)
	fn := reflect.MakeFunc(fnT, func(args []reflect.Value) (results []reflect.Value) {
		deployer, err := NewDeployInvoker(property.SockPath)
		if err != nil {
			return []reflect.Value{reflect.Zero(outType), reflect.ValueOf(err)}
		}

		val := reflect.New(outType).Elem()
		val.Field(1).Set(reflect.ValueOf(deployer))
		return []reflect.Value{val, types.NilError}
	}).Interface()
	return fn
}

func MakeExecDepedency(property *types.DependencyProperty) interface{} {
	tag := ""
	if !property.Require {
		tag = `optional:"true"`
	}
	tag = tag + fmt.Sprintf(` name:"%s"`, property.Value)
	fileds := []reflect.StructField{
		{
			Name:      "OutVal",
			Type:      env.IExecT,
			Tag:       reflect.StructTag(tag),
			Offset:    0,
			Index:     []int{int(0)},
			Anonymous: false,
		},
		{
			Name:      "Out",
			Type:      reflect.TypeOf(fx.Out{}),
			Offset:    1,
			Index:     []int{int(1)},
			Anonymous: true,
		},
	}

	outType := reflect.StructOf(fileds)
	fnT := reflect.FuncOf([]reflect.Type{}, []reflect.Type{outType, types.ErrT}, false)
	fn := reflect.MakeFunc(fnT, func(args []reflect.Value) (results []reflect.Value) {
		exec, err := NewExecInvoker(property.SockPath)
		if err != nil {
			return []reflect.Value{reflect.Zero(outType), reflect.ValueOf(err)}
		}
		val := reflect.New(outType).Elem()
		val.Field(0).Set(reflect.ValueOf(exec))
		return []reflect.Value{val, types.NilError}
	}).Interface()
	return fn
}

func getSvcMap(properties ...*types.DependencyProperty) (map[string]string, error) {
	var svcMap = make(map[string]string)
	for _, p := range properties {
		if !reflect2.IsNil(p.Value) {
			if len(p.Value) > 0 {
				svcMap[p.Name] = p.Value
			}
		}
	}
	return svcMap, nil
}

func convertInjectParams(in reflect.Type, svcMap map[string]string) reflect.Type {
	var inDepTypeFields []reflect.StructField
	offset := uintptr(0)
	if in != nil {
		fieldNum := in.NumField()
		for i := 0; i < fieldNum; i++ {
			field := in.Field(i)
			if !field.Anonymous {
				svcNameKey, found := field.Tag.Lookup(SvcName)
				svcName := ""
				if found {
					var ok bool
					svcName, ok = svcMap[svcNameKey]
					if !ok {
						svcName = uuid.NewString() //use a non exit name, means set the value to nil
					}
				}

				tagVal := fmt.Sprintf(`name:"%s"`, svcName)
				if field.Tag.Get("optional") == "true" {
					tagVal = fmt.Sprintf(`%s optional:"true"`, tagVal)
				}
				newField := reflect.StructField{
					Name:      field.Name,
					PkgPath:   field.PkgPath,
					Type:      field.Type,
					Tag:       reflect.StructTag(tagVal),
					Offset:    offset,
					Index:     []int{int(offset)},
					Anonymous: false,
				}
				inDepTypeFields = append(inDepTypeFields, newField)
				offset++
			}
		}
	}

	inDepTypeFields = append(inDepTypeFields, reflect.StructField{
		Name:      "In",
		Type:      reflect.TypeOf(fx.In{}),
		Offset:    offset,
		Index:     []int{int(offset)},
		Anonymous: true,
	})
	return reflect.StructOf(inDepTypeFields)
}

func collectParams(properties []*types.Property, params interface{}) error {
	value := make(map[string]interface{})
	var err error
	for _, p := range properties {
		value[p.Name], err = GetPropertyValue(p)
		if err != nil {
			return err
		}
	}
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonBytes, params)
	if err != nil {
		return err
	}
	return nil
}

func StartDeployPlugin(deployer env.IDeployer, intParmas *InitParams) error {
	e := gin.Default()
	e.Use(errorHandleMiddleWare())
	e.GET("/pods", func(ctx *gin.Context) {
		pods, err := deployer.Pods(ctx)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		ctx.AbortWithStatusJSON(http.StatusOK, pods)
	})

	e.GET("/statefulset", func(ctx *gin.Context) {
		statefulSet, err := deployer.StatefulSet(ctx)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		ctx.AbortWithStatusJSON(http.StatusOK, statefulSet)
	})

	e.GET("/svc", func(ctx *gin.Context) {
		svc, err := deployer.Svc(ctx)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		ctx.AbortWithStatusJSON(http.StatusOK, svc)
	})

	e.GET("/svcendpoint", func(ctx *gin.Context) {
		svc, err := deployer.SvcEndpoint()
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusOK, svc)
	})

	e.POST("/deploy", func(ctx *gin.Context) {
		err := deployer.Deploy(ctx)
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		ctx.AbortWithStatus(http.StatusOK)
	})

	e.GET("/getconfig", func(ctx *gin.Context) {
		cfg, err := deployer.GetConfig(ctx)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		ctx.AbortWithStatusJSON(http.StatusOK, cfg)
	})

	e.POST("/update", func(ctx *gin.Context) {
		data, err := ioutil.ReadAll(ctx.Request.Body)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		err = deployer.Update(ctx, data)
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		ctx.AbortWithStatus(http.StatusOK)
	})

	e.GET("/params/:params", func(ctx *gin.Context) {
		key, found := ctx.Params.Get("params")
		if !found {
			_ = ctx.Error(fmt.Errorf("key not found"))
			return
		}

		p, err := deployer.Param(key)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		ctx.JSON(http.StatusOK, p)
	})

	shutdownCh := make(types.Shutdown)
	go func() {
		err := e.RunUnix(intParmas.SockPath)
		if err != nil {
			logDetail.Errorf("listen unix fail %v", err)
			shutdownCh <- struct{}{}
		}
	}()

	go types.CatchSig(context.Background(), shutdownCh)
	respState(COMPLETELOG)
	<-shutdownCh
	logDetail.Infof("gracefully shutdown")
	return nil
}

func StartExecPlugin(exec env.IExec, intParmas *InitParams) error {
	e := gin.Default()
	e.Use(errorHandleMiddleWare())
	e.GET("/params/:params", func(ctx *gin.Context) {
		key, found := ctx.Params.Get("params")
		if !found {
			_ = ctx.Error(fmt.Errorf("key not found"))
			return
		}

		p, err := exec.Param(key)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		ctx.JSON(http.StatusOK, p)
	})

	shutdownCh := make(types.Shutdown)
	go func() {
		err := e.RunUnix(intParmas.SockPath)
		if err != nil {
			logDetail.Errorf("listen unix fail %v", err)
			shutdownCh <- struct{}{}
		}
	}()

	go types.CatchSig(context.Background(), shutdownCh)
	respState(COMPLETELOG)
	<-shutdownCh
	logDetail.Infof("gracefully shutdown")
	return nil
}

func errorHandleMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Errors != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			c.Writer.Write([]byte(strings.Join(c.Errors.Errors(), ","))) //nolint
			return
		}
		c.Writer.WriteHeader(http.StatusOK)
	}
}
