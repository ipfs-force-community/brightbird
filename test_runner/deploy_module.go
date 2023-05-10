package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"strconv"
	"syscall"

	"github.com/hunjixin/brightbird/types"

	"github.com/google/uuid"

	"github.com/hunjixin/brightbird/env/plugin"

	"github.com/hunjixin/brightbird/env"

	"github.com/hunjixin/brightbird/repo"

	"github.com/hunjixin/brightbird/fx_opt"
	"go.uber.org/fx"
)

type InitedNode map[string]string

func DeployFLow(ctx context.Context, initedNode InitedNode, pluginRepo repo.IPluginService, pluginStore, testId string, k8sEnvParams *env.K8sInitParams, deployers []*types.DeployNode) fx_opt.Option {
	var opts []fx_opt.Option
	for _, dep := range deployers {
		deployPlugin, err := pluginRepo.GetPlugin(ctx, dep.Name, dep.Version)
		if err != nil {
			opts = append(opts, fx_opt.Error(fmt.Errorf("unable to get plugin %s %s %w", dep.Name, dep.Version, err)))
			break
		}

		for _, dep := range dep.Dependencies {
			sockPath := initedNode[dep.Value]
			dep.SockPath = sockPath
		}

		codeVersionProp := plugin.FindCodeVersionProperties(dep.Properties)
		instanceName := dep.InstanceName.Value
		tmpFName := path.Join(os.TempDir(), fmt.Sprintf("%s_%s_%s.sock", testId, instanceName, uuid.New().String()))
		params := &plugin.InitParams{
			K8sInitParams: *k8sEnvParams,
			BaseConfig: env.BaseConfig{
				CodeVersion:  codeVersionProp.Value.(string),
				InstanceName: instanceName,
			},
			SockPath:     tmpFName,
			Dependencies: dep.Dependencies,
			Properties:   dep.Properties,
		}
		initedNode[instanceName] = tmpFName

		newFn, err := makeDeployPluginSetupFunc(params, path.Join(pluginStore, deployPlugin.Path), instanceName)
		if err != nil {
			opts = append(opts, fx_opt.Error(err))
			break
		}

		opts = append(opts, fx_opt.Annotate(struct {
			ty  reflect.Type
			tag string
		}{
			ty:  MasterDeployInvokerT,
			tag: instanceName,
		}, newFn))
	}
	return fx_opt.Options(opts...)
}

func ExecFlow(ctx context.Context, initedNode InitedNode, pluginRepo repo.IPluginService, pluginStore, testId string, k8sEnvParams *env.K8sInitParams, testItems []*types.TestItem) fx_opt.Option {
	var opts []fx_opt.Option
	var invokeFields []reflect.StructField
	for index, dep := range testItems {
		execPlugin, err := pluginRepo.GetPlugin(ctx, dep.Name, dep.Version)
		if err != nil {
			opts = append(opts, fx_opt.Error(fmt.Errorf("unable to get plugin %s %s %w", dep.Name, dep.Version, err)))
			break
		}

		instanceName := dep.InstanceName.Value
		tmpFName := path.Join(os.TempDir(), fmt.Sprintf("%s_%s_%s.sock", testId, instanceName, uuid.New().String()))
		for _, dep := range dep.Dependencies {
			sockPath := initedNode[dep.Value]
			dep.SockPath = sockPath
		}
		params := &plugin.InitParams{
			K8sInitParams: *k8sEnvParams,
			BaseConfig: env.BaseConfig{
				CodeVersion:  "", //exec have no code version
				InstanceName: instanceName,
			},
			SockPath:     tmpFName,
			Dependencies: dep.Dependencies,
			Properties:   dep.Properties,
		}
		initedNode[instanceName] = tmpFName

		newFn, err := makeExecPluginSetupFunc(params, path.Join(pluginStore, execPlugin.Path), instanceName)
		if err != nil {
			opts = append(opts, fx_opt.Error(err))
			break
		}

		opts = append(opts, fx_opt.Annotate(struct {
			ty  reflect.Type
			tag string
		}{
			ty:  MasterExecInvokerT,
			tag: instanceName,
		}, newFn))

		invokeFields = append(invokeFields, reflect.StructField{
			Name:      "N" + strconv.Itoa(index), //just placeorder
			Type:      MasterExecInvokerT,
			Tag:       reflect.StructTag(fmt.Sprintf(`name:"%s"`, instanceName)),
			Offset:    0,
			Index:     []int{index},
			Anonymous: false,
		})
	}
	invokeFields = append(invokeFields, reflect.StructField{
		Name:      "In",
		Type:      reflect.TypeOf(fx.In{}),
		Offset:    1,
		Index:     []int{len(testItems)},
		Anonymous: true,
	})
	invokeType := reflect.StructOf(invokeFields)
	fnT := reflect.FuncOf([]reflect.Type{invokeType}, []reflect.Type{}, false)
	fn := reflect.MakeFunc(fnT, func(args []reflect.Value) (results []reflect.Value) {
		return nil
	}).Interface()
	opts = append(opts, fx_opt.Override(fx_opt.NextInvoke(), fn))
	return fx_opt.Options(opts...)
}

func makeDeployPluginSetupFunc(params *plugin.InitParams, pluginPath, instanceName string) (interface{}, error) {
	// prepare params
	//start exec
	//return interface type
	retType := reflect.StructOf([]reflect.StructField{
		{
			Name:      "Instance", //just placeorder
			Type:      MasterDeployInvokerT,
			Tag:       reflect.StructTag(fmt.Sprintf(`name:"%s"`, instanceName)),
			Offset:    0,
			Index:     []int{0},
			Anonymous: false,
		},
		{
			Name:      "In",
			Type:      reflect.TypeOf(fx.Out{}),
			Offset:    1,
			Index:     []int{1},
			Anonymous: true,
		},
	})

	depedencyT, err := makeDedenciesType(params.Dependencies)
	if err != nil {
		return nil, err
	}
	fnT := reflect.FuncOf([]reflect.Type{depedencyT}, []reflect.Type{retType, types.ErrT}, false)
	return reflect.MakeFunc(fnT, func(args []reflect.Value) (results []reflect.Value) {
		process, err := runPluginAndWaitForReady(params, pluginPath)
		if err != nil {
			return []reflect.Value{reflect.Zero(retType), reflect.ValueOf(err)}
		}

		deployer, err := SetupMasterDeploy(params.SockPath, process) //used to give argument
		if err != nil {
			return []reflect.Value{reflect.Zero(retType), reflect.ValueOf(err)}
		}

		retVal := reflect.New(retType).Elem()
		retVal.Field(0).Set(reflect.ValueOf(deployer))
		return []reflect.Value{retVal, types.NilError}
	}).Interface(), nil

}

func makeDedenciesType(depedencies []*types.DependencyProperty) (reflect.Type, error) {
	depedenceiesFields := []reflect.StructField{
		{
			Name:      "In",
			Type:      reflect.TypeOf(fx.In{}),
			Offset:    0,
			Index:     []int{0},
			Anonymous: true,
		},
	}

	for index, depedepedency := range depedencies {
		tagVal := ""
		if len(depedepedency.Value) > 0 {
			tagVal = fmt.Sprintf(`name:"%s"`, depedepedency.Value)
		}

		if !depedepedency.Require {
			tagVal = fmt.Sprintf(`%s optional:"true"`, tagVal)
		}
		newFiled := reflect.StructField{
			Name:      "N" + strconv.Itoa(index),
			PkgPath:   "",
			Type:      MasterDeployInvokerT,
			Tag:       reflect.StructTag(tagVal),
			Offset:    uintptr(index + 1),
			Index:     []int{int(index + 1)},
			Anonymous: false,
		}
		switch depedepedency.Type {
		case types.Deploy:
			newFiled.Type = MasterDeployInvokerT
		case types.TestExec:
			newFiled.Type = MasterExecInvokerT
		default:
			return nil, fmt.Errorf("not support plugin type %s", depedepedency.Name)
		}

		depedenceiesFields = append(depedenceiesFields, newFiled)
	}
	return reflect.StructOf(depedenceiesFields), nil
}

var MasterDeployInvokerT = reflect.TypeOf(&MasterDeployInvoker{})

type MasterDeployInvoker struct {
	*plugin.DeployInvoker
	process *os.Process
}

func SetupMasterDeploy(sockPath string, process *os.Process) (*MasterDeployInvoker, error) {
	invoker, err := plugin.NewDeployInvoker(sockPath)
	if err != nil {
		return nil, err
	}
	return &MasterDeployInvoker{
		DeployInvoker: invoker,
		process:       process,
	}, nil
}

func (serve *MasterDeployInvoker) Stop(ctx context.Context) error {
	return serve.process.Kill()
}

func makeExecPluginSetupFunc(params *plugin.InitParams, pluginPath, instanceName string) (interface{}, error) {
	// prepare params
	// start exec
	// return interface type
	retType := reflect.StructOf([]reflect.StructField{
		{
			Name:      "Instance", //just placeorder
			Type:      MasterExecInvokerT,
			Tag:       reflect.StructTag(fmt.Sprintf(`name:"%s"`, instanceName)),
			Offset:    0,
			Index:     []int{0},
			Anonymous: false,
		},
		{
			Name:      "In",
			Type:      reflect.TypeOf(fx.Out{}),
			Offset:    1,
			Index:     []int{1},
			Anonymous: true,
		},
	})

	depedencyT, err := makeDedenciesType(params.Dependencies)
	if err != nil {
		return nil, err
	}
	fnT := reflect.FuncOf([]reflect.Type{depedencyT}, []reflect.Type{retType, types.ErrT}, false)
	return reflect.MakeFunc(fnT, func(args []reflect.Value) (results []reflect.Value) {
		process, err := runPluginAndWaitForReady(params, pluginPath)
		if err != nil {
			return []reflect.Value{reflect.Zero(retType), reflect.ValueOf(err)}
		}

		exec, err := SetupMasterExec(params.SockPath, process) //used to give argument
		if err != nil {
			return []reflect.Value{reflect.Zero(retType), reflect.ValueOf(err)}
		}
		retVal := reflect.New(retType).Elem()
		retVal.Field(0).Set(reflect.ValueOf(exec))
		return []reflect.Value{retVal, types.NilError}
	}).Interface(), nil
}

var MasterExecInvokerT = reflect.TypeOf(&MasterExecInvoker{})

type MasterExecInvoker struct {
	*plugin.ExecInvoker
	process *os.Process
}

func SetupMasterExec(sockPath string, process *os.Process) (*MasterExecInvoker, error) {
	invoker, err := plugin.NewExecInvoker(sockPath)
	if err != nil {
		return nil, err
	}
	return &MasterExecInvoker{
		ExecInvoker: invoker,
		process:     process,
	}, nil
}

func (serve *MasterExecInvoker) Stop(ctx context.Context) error {
	return serve.process.Signal(syscall.SIGQUIT)
}

func runPluginAndWaitForReady(params *plugin.InitParams, pluginPath string) (*os.Process, error) {
	// standard input, standard output, and standard error.
	stdInR, stdInW, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	stdOutR, stdOutW, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	//todo close pipe
	r := bufio.NewReader(io.TeeReader(stdOutR, os.Stdout))
	readyCh := make(chan struct{})
	errCh := make(chan error)
	go func() {
		for {
			data, err := r.ReadString('\n')
			if err != nil {
				errCh <- err
				return
			}
			cmd, val, isCmd := plugin.ReadCMD(data)
			if isCmd {
				switch cmd {
				case plugin.CMDSTATEPREFIX:
					if val == plugin.COMPLETELOG {
						readyCh <- struct{}{}
					}
				case plugin.CMDERRORREFIX:
					errCh <- fmt.Errorf("get error form plugin %s %s", params.InstanceName, val)
				case plugin.CMDVALPREFIX:
				}
			}
		}
	}()

	process, err := os.StartProcess(pluginPath, []string{pluginPath}, &os.ProcAttr{
		Env:   os.Environ(),
		Files: []*os.File{stdInR, stdOutW, os.Stderr},
	})
	if err != nil {
		fmt.Print(err.Error())
		return nil, err
	}
	//write response
	initData, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(initData))
	_, err = stdInW.Write(initData)
	if err != nil {
		return nil, err
	}
	_, err = stdInW.Write([]byte{'\n'})
	if err != nil {
		return nil, err
	}
	//wait for specific log
	select {
	case <-readyCh:
		return process, nil
	case err = <-errCh:
		return nil, err
	}
}
