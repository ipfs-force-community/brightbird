package plugin

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/hunjixin/brightbird/env"

	"github.com/hunjixin/brightbird/types"
)

func ParseParams(params reflect.Type) (types.PluginParams, error) {
	pluginParams := types.PluginParams{}
	numFields := params.NumField()
	var svcProperties []types.DependencyProperty
	for i := 0; i < numFields; i++ {
		field := params.Field(i)
		if field.Name == "Params" {
			configProperties, err := ParserProperties(field.Type)
			if err != nil {
				return types.PluginParams{}, err
			}
			pluginParams.Properties = configProperties
		} else {
			optional := field.Tag.Get(Optional)
			svcName := field.Tag.Get(SvcName)
			if len(svcName) == 0 {
				continue
			}
			if field.Type == env.IDeployerT {
				svcProperties = append(svcProperties, types.DependencyProperty{
					Name:        svcName,
					Type:        types.Deploy,
					Description: "",
					Require:     optional != "true",
				})
			} else if field.Type == env.IExecT {
				svcProperties = append(svcProperties, types.DependencyProperty{
					Name:        svcName,
					Type:        types.TestExec,
					Description: "",
					Require:     optional != "true",
				})
			} else {
				return types.PluginParams{}, errors.New("unsupport plugin type")
			}
		}
	}
	pluginParams.Dependencies = svcProperties
	return pluginParams, nil
}

func ParserProperties(configT reflect.Type) ([]types.Property, error) {
	configFieldsNum := configT.NumField()
	var properties []types.Property
	for j := 0; j < configFieldsNum; j++ {
		field := configT.Field(j)
		if field.Anonymous {
			embedProperties, err := ParserProperties(field.Type)
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
		properties = append(properties, types.Property{
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
	return "", fmt.Errorf("types %s not support", val.String())
}
