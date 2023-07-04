package env

import (
	"encoding/json"
	"fmt"

	"github.com/hunjixin/brightbird/types"
)

type GlobalParams struct {
	LogLevel         string
	CustomProperties json.RawMessage
}

type NodeContext struct {
	Input  json.RawMessage
	OutPut json.RawMessage
}

type EnvContext struct {
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
	StatefulSetName string         `json:"statefulSetName"`
	ConfigMapName   string         `json:"configMapName"`
	SVCName         string         `json:"svcName"`
	SvcEndpoint     types.Endpoint `json:"svcEndpoint"`
}
