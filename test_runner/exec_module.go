package main

import (
	"fmt"
	"reflect"
	"runtime/debug"

	"github.com/hunjixin/brightbird/fx_opt"
	"github.com/hunjixin/brightbird/types"
)

func GenInvokeExec(plugin *types.PluginDetail, testItem *types.TestItem) (interface{}, error) {
	svcMap, err := getSvcMap(testItem.SvcProperties...)
	if err != nil {
		return nil, err
	}

	newInStruct := convertInjectParams(plugin.Param, svcMap)
	//paramsT can't be pointer type
	fnT := reflect.FuncOf([]reflect.Type{types.CtxT, newInStruct}, []reflect.Type{types.ErrT}, false)
	return reflect.MakeFunc(fnT, func(args []reflect.Value) (results []reflect.Value) {
		defer func() {
			if r := recover(); r != nil {
				log.Info("stacktrace from panic:" + string(debug.Stack()))
				results = []reflect.Value{reflect.ValueOf(fmt.Errorf("invoke exec plugin %v", r))}
			}
		}()
		//json must use pointer,
		//1. create a pointer value
		//2. set injected values
		//3. unmarshal pointer value
		log.Infof("start to exec %s", plugin.Name)
		ptrParams := reflect.New(plugin.Param).Elem()
		//apply new struct values
		for j := 0; j < newInStruct.NumField(); j++ {
			field := newInStruct.Field(j)
			fieldName := field.Name
			if fieldName == "Params" {
				val := reflect.New(field.Type)
				err := collectParams(testItem.Properties, val.Interface())
				if err != nil {
					return []reflect.Value{reflect.ValueOf(err)}
				}
				ptrParams.FieldByName(fieldName).Set(val.Elem())
			} else {
				ptrParams.FieldByName(fieldName).Set(args[1].FieldByName(fieldName))
			}
		}
		args[1] = ptrParams
		return plugin.Fn.Call(args)
	}).Interface(), nil
}

func ExecFlow(pluginStore *types.PluginStore, testItems []*types.TestItem) fx_opt.Option {
	var opts []fx_opt.Option
	for _, dep := range testItems {
		pluginInfo, err := pluginStore.GetPlugin(dep.Name)
		if err != nil {
			opts = append(opts, fx_opt.Error(err))
			break
		}
		fn, err := GenInvokeExec(pluginInfo, dep)
		if err != nil {
			opts = append(opts, fx_opt.Error(err))
			break
		}
		opts = append(opts, fx_opt.Override(fx_opt.NextInvoke(), fn))
	}
	return fx_opt.Options(opts...)
}
