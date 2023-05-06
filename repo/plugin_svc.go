package repo

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/hunjixin/brightbird/types"
)

type IPluginService interface {
	DeployPlugins(context.Context) ([]types.PluginOut, error)
	ExecPlugins(context.Context) ([]types.PluginOut, error)
	GetByName(context.Context, string) (*types.PluginOut, error)
}

type PluginSvc struct {
	deployPluginStore DeployPluginStore
	execPluginStore   ExecPluginStore
}

func NewPluginSvc(deployPluginStore DeployPluginStore, execPluginStore ExecPluginStore) *PluginSvc {
	return &PluginSvc{deployPluginStore: deployPluginStore, execPluginStore: execPluginStore}
}

func (p *PluginSvc) DeployPlugins(ctx context.Context) ([]types.PluginOut, error) {
	var deployPlugins []types.PluginOut
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
	sort.Slice(deployPlugins, func(i, j int) bool {
		return deployPlugins[i].Name > deployPlugins[j].Name
	})
	return deployPlugins, nil
}

func (p *PluginSvc) ExecPlugins(ctx context.Context) ([]types.PluginOut, error) {
	var deployPlugins []types.PluginOut
	err := p.execPluginStore.Each(func(detail *types.PluginDetail) error {
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
	sort.Slice(deployPlugins, func(i, j int) bool {
		return deployPlugins[i].Name > deployPlugins[j].Name
	})
	return deployPlugins, nil
}

func (p *PluginSvc) GetByName(ctx context.Context, name string) (*types.PluginOut, error) {
	var deployPlugins *types.PluginOut
	err := p.deployPluginStore.Each(func(detail *types.PluginDetail) error {
		pluginOut, err := getPluginOutput(detail)
		if err != nil {
			return err
		}
		if pluginOut.Name == name {
			deployPlugins = &pluginOut
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = p.execPluginStore.Each(func(detail *types.PluginDetail) error {
		pluginOut, err := getPluginOutput(detail)
		if err != nil {
			return err
		}
		if pluginOut.Name == name {
			deployPlugins = &pluginOut
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if deployPlugins == nil {
		return nil, fmt.Errorf("plugin %s not found", name)
	}
	return deployPlugins, nil
}

func getPluginOutput(detail *types.PluginDetail) (types.PluginOut, error) {
	var pluginOut = types.PluginOut{}
	pluginOut.PluginInfo = *detail.PluginInfo
	numFields := detail.Param.NumField()

	var svcProperties []types.Property
	for i := 0; i < numFields; i++ {
		field := detail.Param.Field(i)
		if field.Name == "Params" {
			configProperties, err := parserProperties(field.Type)
			if err != nil {
				return types.PluginOut{}, err
			}
			pluginOut.Properties = configProperties
		} else {
			svcName := field.Tag.Get(types.SvcName)
			if len(svcName) == 0 {
				continue
			}
			svcType := strings.TrimRight(strings.TrimLeft(field.Type.Name(), "I"), "Deployer")
			svcProperties = append(svcProperties, types.Property{
				Name:        svcName,
				Type:        svcType,
				Description: "",
			})
		}
	}
	pluginOut.SvcProperties = svcProperties
	pluginOut.Out = &types.Property{
		Name:        "Out",
		Type:        "string",
		Description: "",
	}
	return pluginOut, nil
}

func parserProperties(configT reflect.Type) ([]types.Property, error) {
	configFieldsNum := configT.NumField()
	var properties []types.Property
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
	return "", fmt.Errorf("types %t not support", val.String())
}
