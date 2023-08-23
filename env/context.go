package env

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs-force-community/brightbird/types"
)

// logLevel
type GlobalParams map[string]interface{}

func (global GlobalParams) GetProperty(key string, val interface{}) error {
	property, ok := global[key]
	if !ok {
		return fmt.Errorf("key %s not found", key)
	}

	data, err := json.Marshal(property)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, val)
}

func (global GlobalParams) GetStringProperty(key string) (string, error) {
	propertyVal, ok := global[key]
	if !ok {
		return "", fmt.Errorf("key %s not found", key)
	}

	val, ok := propertyVal.(string)
	if !ok {
		return "", fmt.Errorf("value %v is not string", propertyVal)
	}
	return val, nil
}

type NodeContext struct {
	Input  json.RawMessage
	OutPut json.RawMessage
}

type EnvContext struct { //nolint
	Global         GlobalParams
	Nodes          map[string]*NodeContext
	CurrentContext string
}

func (envCtx EnvContext) Current() *NodeContext {
	return envCtx.Nodes[envCtx.CurrentContext]
}

func (envCtx EnvContext) GetNode(name string) (*NodeContext, error) {
	val, ok := envCtx.Nodes[name]
	if !ok {
		return nil, fmt.Errorf("node %s not found", name)
	}
	return val, nil
}

type CommonDeployParams struct {
	BaseConfig
	DeployName      string         `json:"deployName"`
	StatefulSetName string         `json:"statefulSetName"`
	ConfigMapName   string         `json:"configMapName"`
	SVCName         string         `json:"svcName"`
	SvcEndpoint     types.Endpoint `json:"svcEndpoint"`
}

type EmptyReturn struct{}
