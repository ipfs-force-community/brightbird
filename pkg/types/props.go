package types

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

type ExecNode struct {
	// the name for this test flow
	// required: true
	// min length: 3
	Name string `json:"name" yaml:"name"`
	// the name for this test flow
	// required: true
	// min length: 3
	InstanceName string `json:"instanceName" yaml:"instanceName"`

	PluginType PluginType `json:"pluginType" yaml:"pluginType"`
	// the version for this test flow
	// required: true
	// min length: 3
	Version string `json:"version" yaml:"version"`

	Input  json.RawMessage `json:"input" yaml:"input"`
	Output json.RawMessage `json:"output" yaml:"output"`
}

func (h *ExecNode) UnmarshalYAML(value *yaml.Node) error {
	for i := 0; i < len(value.Content); i += 2 {
		switch value.Content[i].Value {
		case "name":
			h.Name = value.Content[i+1].Value
		case "instanceName":
			h.InstanceName = value.Content[i+1].Value
		case "pluginType":
			h.PluginType = PluginType(value.Content[i+1].Value)
		case "version":
			h.Version = value.Content[i+1].Value
		case "input":
			var anyV interface{}
			if err := value.Content[i+1].Decode(&anyV); err != nil {
				return err
			}
			jsonRaw, err := json.Marshal(anyV)
			if err != nil {
				return err
			}
			h.Input = jsonRaw
		case "output":
			var anyV interface{}
			if err := value.Content[i+1].Decode(&anyV); err != nil {
				return err
			}
			jsonRaw, err := json.Marshal(anyV)
			if err != nil {
				return err
			}
			h.Output = jsonRaw
		}
	}

	return nil
}
