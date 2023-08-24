package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MyString repretation string
// swagger:model myString
type MyString string

// ObjectId
// swagger:model objectId
type ObjectId primitive.ObjectID

// APIError
// swagger:model apiError
type APIError struct {
	Message string `json:"message"`
}

// UpdateGroupRequest
// swagger:model updateGroupRequest
type UpdateGroupRequest struct {
	Name        string `json:"name"`
	IsShow      bool   `json:"isShow"`
	Description string `json:"description"`
}

// GroupResp
// swagger:model groupResp
type GroupResp struct {
	*Group
	TestFlowCount int `json:"testFlowCount"`
}

// ListGroupResp
// swagger:model listGroupResp
type ListGroupResp []GroupResp

// CountGroupRequest
// swagger:model countGroupRequest
type CountGroupRequest struct {
	ID   *string `form:"id"`
	Name *string `form:"name"`
}

// UpdateJobRequest
// swagger:model updateJobRequest
type UpdateJobRequest struct {
	TestFlowId  primitive.ObjectID `json:"testFlowId"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	//cron job params
	CronJobParams
	Versions map[string]string `json:"versions"`
}

// JobDetailResp
// swagger:model jobDetailResp
type JobDetailResp struct {
	Job
	TestFlowName string `json:"testFlowName"`
	GroupName    string `json:"groupName"`
}

// JobNextNReq
// swagger:parameters jobNextNReq
type JobNextNReq struct {
	ID string `form:"id" json:"id"`
	N  int    `form:"n" json:"n"`
}

// CountJobRequest
// swagger:parameters countJobRequest
type CountJobRequest struct {
	// id of job
	ID *string `form:"id" json:"id"`
	// name of job
	Name *string `form:"name" json:"name"`
}

// ListJobResp
// swagger:model listJobResp
type ListJobResp []Job

// StringArray
// swagger:model stringArr
type StringArray []string

// Int64Array
// swagger:model int64Arr
type Int64Array []int64

// LogListResp
// swagger:model logListResp
type LogListResp []string

// ListTasksParams
// swagger:parameters listTasksParams
type ListTasksParams struct {
	// id of job
	JobID string `form:"jobId" json:"jobId"` //todo use objectid directly issue https://github.com/gin-gonic/gin/issues/2447
	// task state
	State []State `form:"state" json:"state"`
	// createtime of task
	CreateTime *int64 `form:"createTime" json:"createTime"`
}

// GetTaskReq
// swagger:parameters getTaskReq
type GetTaskReq struct {
	TestID *string `form:"testID" json:"testID"`
	ID     *string `form:"ID" json:"ID"`
}

// ListTasksReq
// swagger:parameters listTasksReq
type ListTasksReq struct { //todo https://github.com/go-swagger/go-swagger/issues/2802

	// pageNum
	//
	// required: true
	// in: query
	PageNum int64 `form:"pageNum" json:"pageNum" binding:"required,gte=1"`

	// pageSize
	//
	// required: true
	// in: query
	PageSize int64 `form:"pageSize" json:"pageSize" binding:"required,gte=1"`

	ListTasksParams
}

// ListTasksResp
// swagger:model listTasksResp
type ListTasksResp struct { //todo https://github.com/go-swagger/go-swagger/issues/2802
	Total   int64   `json:"total"`
	Pages   int64   `json:"pages"`
	PageNum int64   `json:"pageNum"`
	List    []*Task `json:"list"`
}

// ListInGroupParams
// swagger:parameters listInGroupParams
type ListInGroupParams struct {
	//id of group
	GroupID *string `form:"groupId" json:"groupId"`
	//name of testflow
	Name *string `form:"name" json:"name"`
}

// ListInGroupRequest
// swagger:parameters listInGroupRequest
type ListInGroupRequest struct { //todo https://github.com/go-swagger/go-swagger/issues/2802
	PageNum  int64 `form:"pageNum" json:"pageNum" binding:"required,gte=1"`
	PageSize int64 `form:"pageSize" json:"pageSize" binding:"required,gte=1"`
	ListInGroupParams
}

// GetTestFlowRequest
// swagger:parameters getTestFlowRequest
type GetTestFlowRequest struct {
	//id of testflow
	ID *string `form:"id" json:"id"`
	//name of testflow
	Name *string `form:"name" json:"name"`
}

// CountTestFlowRequest
// swagger:parameters countTestFlowRequest
type CountTestFlowRequest struct {
	//id of group
	GroupID *string `form:"groupId" json:"groupId"`
	//name of testflow
	Name *string `form:"name" json:"name"`
}

// ListTestFlowResp
// swagger:model listTestFlowResp
type ListTestFlowResp struct { //todo https://github.com/go-swagger/go-swagger/issues/2802
	Total   int64      `json:"total"`
	Pages   int64      `json:"pages"`
	PageNum int64      `json:"pageNum"`
	List    []TestFlow `json:"list"`
}

// ChangeTestflowGroupRequest
// swagger:model changeTestflowGroupRequest
type ChangeTestflowGroupRequest struct {
	// Group to change
	//
	// required: true
	// in: body
	GroupID primitive.ObjectID `json:"groupId"`
	// testflow id
	//
	// required: true
	// in: body
	TestflowIDs []primitive.ObjectID `json:"testflowIds"`
}

// ChangeGroupRequest
// swagger:parameters changeGroupRequest
type ChangeGroupRequest struct {
	// update to submit
	//
	// required: true
	// in: body
	Body ChangeTestflowGroupRequest
}

// DeletePluginByVersionReq
// swagger:parameters deletePluginByVersionReq
type DeletePluginByVersionReq struct {
	// id of plugin
	//
	// required: true
	ID string `form:"id" json:"id" binding:"required"`
	// specific plugin version
	//
	// required: true
	Version string `form:"version" json:"version" binding:"required"`
}

// DeletePluginReq
// swagger:parameters deletePluginReq
type DeletePluginReq struct {
	// id of plugin
	//
	// required: true
	ID string `form:"id" json:"id" binding:"required"`
}

// PodLogReq
// swagger:parameters podLogReq
type PodLogReq struct {
	// testid of task
	//
	// required: true
	TestID string `form:"testID" json:"testID" binding:"required"`
	// pod name
	//
	// required: true
	PodName string `form:"podName" json:"podLog" binding:"required"`
}

type StepState string

const (
	StepNotRunning StepState = "notrunning" //nolint
	StepRunning    StepState = "running"    //nolint
	StepSuccess    StepState = "success"    //nolint
	StepFail       StepState = "fail"       //nolint
)

// StepLog
// swagger:model stepLog
type StepLog struct {
	Name         string    `json:"name"`
	InstanceName string    `json:"instanceName"`
	State        StepState `json:"state"`
	Logs         []string  `json:"logs"`
}

// LogResp
// swagger:model logResp
type LogResp struct {
	PodName string     `json:"podName"`
	Steps   []*StepLog `json:"steps"`
	Logs    []string   `json:"logs"`
}

// CopyTestflow
// swagger:model copyTestflow
type CopyTestflow struct {
	ID   primitive.ObjectID `json:"id" form:"id"`
	Name string             `json:"name" form:"name"`
}
