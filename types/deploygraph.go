package types

import "encoding/json"

type DeployNode struct {
	Name     string
	Type     string
	Annotate string
	Params   json.RawMessage
}

type TestItem struct {
	Name   string
	Params json.RawMessage
}
type TestFlow struct {
	Nodes []DeployNode
	Cases []TestItem
}
