package main

import (
	"context"
	"fmt"
	"github.com/hunjixin/brightbird/types"
	"reflect"
	"strings"
)

// PluginOut
// swagger:model pluginOut
type PluginOut struct {
	types.PluginInfo
	Properties    []Property
	IsAnnotateOut bool
	SvcProperties []Property
	Out           *Property
}

// Property Property
// swagger:model property
type Property struct {
	Name        string
	Type        string
	Description string
}

type IPluginService interface {
	Plugins(context.Context) ([]PluginOut, error)
}

type PluginSvc struct {
	deployPluginStore DeployPluginStore
}

func NewPluginSvc() IPluginService {
	return &PluginSvc{}
}

func (p *PluginSvc) Plugins(ctx context.Context) ([]PluginOut, error) {
	var deployPlugins []PluginOut
	err := p.deployPluginStore.Each(func(detail *types.PluginDetail) error {
		pluginOut, err := getPluginOutput(detail)
		if err != nil {
			return err
		}
		deployPlugins = append(deployPlugins, pluginOut)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return deployPlugins, nil
}

func getPluginOutput(detail *types.PluginDetail) (PluginOut, error) {
	var pluginOut = PluginOut{}
	isAnnotateOut := types.IsAnnotateOut(reflect.New(detail.Param).Elem().Interface())
	pluginOut.IsAnnotateOut = isAnnotateOut
	pluginOut.PluginInfo = *detail.PluginInfo
	numFields := detail.Param.NumField()

	var svcProperties []Property
	for i := 0; i < numFields; i++ {
		field := detail.Param.Field(i)
		if field.Name == "Params" {
			configProperties, err := parserProperties(field.Type)
			if err != nil {
				return PluginOut{}, err
			}
			pluginOut.Properties = configProperties
		} else {
			svcName := field.Tag.Get(types.SvcName)
			if len(svcName) == 0 {
				continue
			}
			svcType := strings.TrimRight(strings.TrimLeft(field.Type.Name(), "I"), "Deployer")
			svcProperties = append(svcProperties, Property{
				Name:        svcName,
				Type:        svcType,
				Description: "",
			})
		}
	}
	pluginOut.SvcProperties = svcProperties
	if isAnnotateOut {
		outType := detail.Fn.Type().Out(0)
		pluginOut.Out = &Property{
			Name:        "Out",
			Type:        strings.TrimRight(strings.TrimLeft(outType.Name(), "I"), "Deployer"),
			Description: "",
		}
	}
	return pluginOut, nil
}

func parserProperties(configT reflect.Type) ([]Property, error) {
	configFieldsNum := configT.NumField()
	var properties []Property
	for j := 0; j < configFieldsNum; j++ {
		field := configT.Field(j)
		if field.Anonymous {
			embedProperties, err := parserProperties(field.Type)
			if err != nil {
				return nil, err
			}
			properties = append(properties, embedProperties...)
			continue
		}

		fieldName := getFieldJsonName(field)
		if fieldName == "-" || fieldName == "" {
			continue
		}
		typeName, err := mapType(field.Type.Kind())
		if err != nil {
			return nil, fmt.Errorf("field %s has unspport type %w", fieldName, err)
		}
		properties = append(properties, Property{
			Name: fieldName,
			Type: typeName,
		})
	}
	return properties, nil
}

func getFieldJsonName(field reflect.StructField) string {
	fieldName := field.Name
	jsonTag := field.Tag.Get("json")
	jsonFlags := strings.Split(jsonTag, ",")
	if val := strings.TrimSpace(jsonFlags[0]); len(val) > 0 {
		fieldName = val
	}
	return fieldName
}

func mapType(val reflect.Kind) (string, error) {
	switch val {
	case reflect.Bool:
		return "bool", nil
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Int64: //todo use bignumber
		fallthrough
	case reflect.Uint64: //todo use bignumber
		return "number", nil
	case reflect.Float32:
		return "decimical", nil
	case reflect.Float64: //todo use bigdecimal
		return "decimal", nil
	case reflect.String:
		return "string", nil
	}
	return "", fmt.Errorf("types %t not support", val.String())
}
