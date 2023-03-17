package types

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BaseTime struct {
	/**
	 * 创建时间
	 */
	CreateTime int64 `json:"createTime,string"`

	/**
	 * 最后修改时间
	 */
	ModifiedTime int64 `json:"modifiedTime,string"`
}

// PluginOut
// swagger:model pluginOut
type PluginOut struct {
	ID primitive.ObjectID `bson:"_id" json:"id"`
	BaseTime
	PluginInfo
	Properties    []Property `json:"properties"`
	IsAnnotateOut bool       `json:"isAnnotateOut"`
	SvcProperties []Property `json:"svcProperties"`
	Out           *Property  `json:"out"`
}

// Property Property
// swagger:model property
type Property struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Value       interface{} `json:"value"`
	Require     bool        `json:"require"`
}

type DeployNode struct {
	// the name for this test flow
	// required: true
	// min length: 3
	Name string `json:"name"`

	IsAnnotateOut bool       `json:"isAnnotateOut"`
	Properties    []Property `json:"properties"`
	SvcProperties []Property `json:"svcProperties"`
	Out           *Property  `json:"out"`
}

type TestItem struct {
	// the name for this test flow
	// required: true
	// min length: 3
	Name string `json:"name"`

	Properties []Property `json:"properties"`
}

// TestFlow
// swagger:model testFlow
type TestFlow struct {
	// the name for this test flow
	// required: true
	// min length: 3
	Name string `json:"name"`

	ID primitive.ObjectID `bson:"_id" json:"id"`
	BaseTime
	GroupId string       `json:"groupId"`
	Nodes   []DeployNode `json:"nodes"`
	Cases   []TestItem   `json:"cases"`
}

// Group
// swagger:model group
type Group struct {
	ID primitive.ObjectID `bson:"_id" json:"id"`
	BaseTime
	// the name for this test flow
	// required: true
	// min length: 3
	Name        string `json:"name"`
	IsShow      bool   `json:"isShow"`
	Description string `json:"description"`
}
