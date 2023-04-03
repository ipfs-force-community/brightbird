package types

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PageReq[T any] struct {
	PageNum  int `form:"pageNum"`
	PageSize int `form:"pageSize"`
	Params   T   `json:"params"`
}

type PageResp[T any] struct {
	Total   int `json:"total"`
	Pages   int `json:"pages"`
	PageNum int `json:"pageNum"`
	List    []T `json:"list"`
}

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

	IsAnnotateOut bool        `json:"isAnnotateOut"`
	Properties    []*Property `json:"properties"`
	SvcProperties []*Property `json:"svcProperties"`
	Out           *Property   `json:"out"`
}

type TestItem struct {
	// the name for this test flow
	// required: true
	// min length: 3
	Name          string      `json:"name"`
	Properties    []*Property `json:"properties"`
	SvcProperties []*Property `json:"svcProperties"`
}

// TestFlow
// swagger:model testFlow
type TestFlow struct {
	ID primitive.ObjectID `bson:"_id" json:"id"`
	// the name for this test flow
	// required: true
	// min length: 3
	Name        string             `json:"name"`
	GroupId     primitive.ObjectID `json:"groupId" bson:"groupId"` //provent mongo use Id to id
	Nodes       []*DeployNode      `json:"nodes"`
	Cases       []*TestItem        `json:"cases"`
	Graph       string             `json:"graph"`
	Description string             `json:"description"`

	BaseTime
}

// Group
// swagger:model group
type Group struct {
	ID primitive.ObjectID `bson:"_id" json:"id"`
	// the name for this test flow
	// required: true
	// min length: 3
	Name        string `json:"name"`
	IsShow      bool   `json:"isShow"`
	Description string `json:"description"`

	BaseTime
}

type JobType string

const (
	CronJobType JobType = "cron_job"
)

// Job
// swagger:model job
type Job struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	TestFlowId  primitive.ObjectID `json:"testFlowId"`
	Name        string             `json:"name"`
	JobType     JobType            `json:"jobType"`
	ExecCount   int                `json:"execCount"`
	Description string             `json:"description"`

	Versions map[string]string `json:"versions"` // save a version setting for user job specific
	//cron job params
	CronJobParams

	BaseTime
}

type CronJobParams struct {
	CronExpression string `json:"cronExpression"`
}

type State string

const (
	Init       State = "init"
	Running    State = "running"
	Error      State = "error"
	Successful State = "success"
)

// Task
// swagger:model task
type Task struct {
	ID         primitive.ObjectID `bson:"_id" json:"id"`
	Name       string             `json:"name"`
	PodName    string             `json:"podName"`
	JobId      primitive.ObjectID `json:"jobId"`
	TestFlowId primitive.ObjectID `json:"testFlowId"` //save this field for convience, get from job info is alright
	TestId     TestId             `json:"testId"`
	State      State              `json:"state"`
	Logs       []string           `json:"logs"`
	Versions   map[string]string  `json:"versions"` // save a copy of task flow, but task flow update version information in this running
	BaseTime
}
