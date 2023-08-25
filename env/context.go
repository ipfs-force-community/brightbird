package env

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/ipfs-force-community/brightbird/types"
)

// logLevel
type GlobalParams map[string]string

func (global GlobalParams) GetJSONProperty(key string, val interface{}) error {
	property, ok := global[key]
	if !ok {
		return fmt.Errorf("key %s not found", key)
	}

	return json.Unmarshal([]byte(property), val)
}

func (global GlobalParams) GetProperty(key string) (string, error) {
	propertyVal, ok := global[key]
	if !ok {
		return "", fmt.Errorf("key %s not found", key)
	}
	return propertyVal, nil
}

func (global GlobalParams) GetNumberProperty(key string) (float64, error) {
	propertyVal, ok := global[key]
	if !ok {
		return math.NaN(), fmt.Errorf("key %s not found", key)
	}

	val, err := strconv.ParseFloat(propertyVal, 64)
	if err != nil {
		return math.NaN(), fmt.Errorf("value %v is not json number %w", propertyVal, err)
	}
	return val, nil
}

type NodeContext struct {
	Input  json.RawMessage `json:"input"`
	OutPut json.RawMessage `json:"output"`
}

type EnvContext struct { //nolint
	Global         GlobalParams            `json:"global"`
	Nodes          map[string]*NodeContext `json:"nodes"`
	CurrentContext string                  `json:"currentContext"`
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
