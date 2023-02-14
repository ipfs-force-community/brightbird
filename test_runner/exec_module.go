package main

import (
	"encoding/json"
	"github.com/hunjixin/brightbird/fx_opt"
	"github.com/hunjixin/brightbird/types"
	"reflect"
)

func GenInvokeExec(plugin *types.PluginDetail, paramsJson json.RawMessage) (interface{}, error) {
	svcMap, err := getSvcMap(paramsJson)
	if err != nil {
		return nil, err
	}

	newInStruct := convertInjectParams(plugin.Param, svcMap)
	//paramsT can't be pointer type
	fnT := reflect.FuncOf([]reflect.Type{types.CtxT, newInStruct}, []reflect.Type{types.ErrT}, false)
	return reflect.MakeFunc(fnT, func(args []reflect.Value) (results []reflect.Value) {
		//json must use pointer,
		//1. create a pointer value
		//2. set injected values
		//3. unmarshal pointer value
		mainLog.Infof("start to exec %s", plugin.Name)
		ptrParams := reflect.New(plugin.Param)
		//apply new struct values
		for j := 0; j < newInStruct.NumField(); j++ {
			field := newInStruct.Field(j)
			if !field.Anonymous {
				ptrParams.Elem().FieldByName(field.Name).Set(args[1].FieldByName(field.Name))
			}
		}
		//apply json value
		err := json.Unmarshal(paramsJson, ptrParams.Interface())
		if err != nil {
			return []reflect.Value{reflect.ValueOf(err)}
		}
		args[1] = ptrParams.Elem()
		return plugin.Fn.Call(args)
	}).Interface(), nil
}

func ExecFlow(pluginStore *types.PluginStore, testItems []types.TestItem) fx_opt.Option {
	var opts []fx_opt.Option
	for _, dep := range testItems {
		pluginInfo, err := pluginStore.GetPlugin(dep.Name)
		if err != nil {
			opts = append(opts, fx_opt.Error(err))
			break
		}
		fn, err := GenInvokeExec(pluginInfo, dep.Params)
		if err != nil {
			opts = append(opts, fx_opt.Error(err))
			break
		}
		opts = append(opts, fx_opt.Override(fx_opt.NextInvoke(), fn))
	}
	return fx_opt.Options(opts...)
}
