package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MyString
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

// PodListResp
// swagger:model podListResp
type PodListResp []string

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
	GroupID     primitive.ObjectID   `json:"groupId"`
	TestflowIDs []primitive.ObjectID `json:"testflowIds"`
}
