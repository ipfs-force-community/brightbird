package types

import "encoding/json"

type DeployNode struct {
	// the name for this test flow
	// required: true
	// min length: 3
	Name   string
	Params json.RawMessage
}

type TestItem struct {
	// the name for this test flow
	// required: true
	// min length: 3
	Name   string
	Params json.RawMessage
}

// TestFlow
// swagger:model testFlow
type TestFlow struct {
	// the name for this test flow
	// required: true
	// min length: 3
	Name  string
	Nodes []DeployNode
	Cases []TestItem
}
