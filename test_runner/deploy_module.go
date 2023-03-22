package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/fx_opt"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
	"go.uber.org/fx"
	"reflect"
	"runtime/debug"
)

func DeployFLow(deployers []*types.DeployNode, deployPlugin *types.PluginStore) fx_opt.Option {
	opts := []fx_opt.Option{
		fx_opt.Override(new(types.AdminToken), func(ctx context.Context, k8sEnv *env.K8sEnvDeployer, authDeploy env.IVenusAuthDeployer) (types.AdminToken, error) {
			endpoint := authDeploy.SvcEndpoint()
			if env.Debug {
				var err error
				endpoint, err = k8sEnv.PortForwardPod(ctx, authDeploy.Pods()[0].GetName(), int(authDeploy.Svc().Spec.Ports[0].Port))
				if err != nil {
					return "", err
				}
			}
			authAPIClient, err := jwtclient.NewAuthClient(endpoint.ToHttp())
			if err != nil {
				return "", err
			}

			_, err = authAPIClient.CreateUser(ctx, &auth.CreateUserRequest{
				Name:    "admin",
				Comment: utils.StringPtr("comment admin"),
				State:   0,
			})
			if err != nil {
				return "", err
			}
			adminToken, err := authAPIClient.GenerateToken(ctx, "admin", "admin", "")
			if err != nil {
				return "", err
			}
			return types.AdminToken(adminToken), nil
		}),
	}
	for _, dep := range deployers {
		plugin, err := deployPlugin.GetPlugin(dep.Name)
		if err != nil {
			opts = append(opts, fx_opt.Error(err))
			break
		}

		newFn, resultTag, err := GenInjectFunc(plugin, dep)
		if err != nil {
			opts = append(opts, fx_opt.Error(err))
			break
		}

		opts = append(opts, fx_opt.Annotate(struct {
			ty  reflect.Type
			tag string
		}{
			ty:  plugin.Fn.Type().Out(0),
			tag: resultTag,
		}, newFn))
	}
	return fx_opt.Options(opts...)
}

func getSvcMap(properties ...*types.Property) (map[string]string, error) {
	var svcMap = make(map[string]string)
	for _, p := range properties {
		if val, ok := p.Value.(string); ok && len(val) > 0 {
			svcMap[p.Name] = val
		}
	}
	return svcMap, nil
}

func convertInjectParams(in reflect.Type, svcMap map[string]string) reflect.Type {
	fieldNum := in.NumField()
	var inDepTypeFields []reflect.StructField
	offset := uintptr(0)
	for i := 0; i < fieldNum; i++ {
		field := in.Field(i)
		if !field.Anonymous {
			svcNameKey := field.Tag.Get(types.SvcName)
			svcName := svcMap[svcNameKey]
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
	inDepTypeFields = append(inDepTypeFields, reflect.StructField{
		Name:      "In",
		Type:      reflect.TypeOf(fx.In{}),
		Offset:    offset,
		Index:     []int{int(offset)},
		Anonymous: true,
	})
	return reflect.StructOf(inDepTypeFields)
}
func GenInjectFunc(plugin *types.PluginDetail, depNode *types.DeployNode) (interface{}, string, error) {
	svcMap, err := getSvcMap(append(depNode.Properties, depNode.Out)...)
	if err != nil {
		return nil, "", err
	}
	newInStruct := convertInjectParams(plugin.Param, svcMap)
	//make function type
	isAnnotateOut := types.IsAnnotateOut(reflect.New(plugin.Param).Elem().Interface())
	fnT := plugin.Fn.Type()
	var newOutArgs []reflect.Type
	var outTag string
	{
		//todo opt for more return values ?
		numOut := fnT.NumOut()
		if numOut != 2 {
			return nil, "", fmt.Errorf("return values must be (val, error) format")
		}
		if isAnnotateOut {
			outTag = svcMap[types.OutLabel]
			resultFields := []reflect.StructField{
				{
					Name:      "OutVal",
					Type:      fnT.Out(0),
					Tag:       reflect.StructTag(fmt.Sprintf(`name:"%s"`, outTag)),
					Offset:    0,
					Index:     []int{int(0)},
					Anonymous: false,
				},
			}
			resultFields = append(resultFields, reflect.StructField{
				Name:      "Out",
				Type:      reflect.TypeOf(fx.Out{}),
				Offset:    1,
				Index:     []int{int(1)},
				Anonymous: true,
			})
			resultStruct := reflect.StructOf(resultFields)
			newOutArgs = append(newOutArgs, resultStruct)
			newOutArgs = append(newOutArgs, types.ErrT)
		} else {
			for i := 0; i < numOut; i++ {
				argT := fnT.Out(i)
				newOutArgs = append(newOutArgs, argT)
			}
		}
	}

	newFn := reflect.FuncOf([]reflect.Type{types.CtxT, newInStruct}, newOutArgs, false)
	return reflect.MakeFunc(newFn, func(args []reflect.Value) (vals []reflect.Value) {
		defer func() {
			if r := recover(); r != nil {
				vals = make([]reflect.Value, 2)
				vals[0] = reflect.Zero(newOutArgs[0])
				log.Info("stacktrace from panic:" + string(debug.Stack()))
				vals[1] = reflect.ValueOf(fmt.Errorf("invoke deploy plugin %v", r))
			}
		}()

		log.Infof("start to deploy %s", depNode.Name)
		//convert params
		argT := fnT.In(1)
		dstVal := reflect.New(argT).Elem()
		for j := 0; j < argT.NumField(); j++ {
			field := argT.Field(j)
			fieldName := field.Name
			if !field.Anonymous && len(fieldName) != 0 {
				if fieldName == "Params" {
					val := reflect.New(field.Type)
					err := collectParams(depNode.Properties, val.Interface())
					if err != nil {
						return []reflect.Value{reflect.Zero(newOutArgs[0]), reflect.ValueOf(err)}
					}
					dstVal.FieldByName(fieldName).Set(reflect.ValueOf(val).Elem())
				} else {
					dstVal.FieldByName(fieldName).Set(args[1].FieldByName(fieldName))
				}
			}
		}

		//call plugin
		results := plugin.Fn.Call([]reflect.Value{args[0], dstVal})
		//convert result
		if isAnnotateOut {
			//todo check error result
			if !results[1].IsNil() {
				return []reflect.Value{reflect.Zero(newOutArgs[0]), results[1]}
			}
			destResultVal := reflect.New(newOutArgs[0]).Elem()
			destResultVal.Field(0).Set(results[0])
			return []reflect.Value{destResultVal, results[1]}
		} else {
			return results
		}
	}).Interface(), outTag, nil
}

func collectParams(properties []*types.Property, params interface{}) error {
	value := make(map[string]interface{})
	for _, p := range properties {
		value[p.Name] = p.Value
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
