package types

import "encoding/json"

type DeployNode struct {
	// the name for this test flow
	// required: true
	// min length: 3
	Name   string          `json:"name"`
	Params json.RawMessage `json:"params"`
}

type TestItem struct {
	// the name for this test flow
	// required: true
	// min length: 3
	Name   string          `json:"name"`
	Params json.RawMessage `json:"params"`
}

// TestFlow
// swagger:model testFlow
type TestFlow struct {
	BaseTime
	// the name for this test flow
	// required: true
	// min length: 3
	Name  string       `json:"name"`
	Nodes []DeployNode `json:"nodes"`
	Cases []TestItem   `json:"cases"`
}
