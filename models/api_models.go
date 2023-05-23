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
// swagger:model countJobRequest
type CountJobRequest struct {
	ID   *string `form:"id"`
	Name *string `form:"name"`
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
// swagger:model listTasksParams
type ListTasksParams struct {
	JobId string  `form:"jobId"` //todo use objectid directly issue https://github.com/gin-gonic/gin/issues/2447
	State []State `form:"state"`
}

// ListTasksReq
// swagger:model listTasksReq
type ListTasksReq struct { //todo https://github.com/go-swagger/go-swagger/issues/2802
	PageNum  int64           `form:"pageNum" binding:"required,gte=1"`
	PageSize int64           `form:"pageSize" binding:"required,gte=1"`
	Params   ListTasksParams `form:"params"`
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
// swagger:model listInGroupParams
type ListInGroupParams struct {
	GroupId *string `form:"groupId"`
	Name    *string `form:"name"`
}

// ListInGroupRequest
// swagger:model listInGroupRequest
type ListInGroupRequest struct { //todo https://github.com/go-swagger/go-swagger/issues/2802
	PageNum  int64             `form:"pageNum" binding:"required,gte=1"`
	PageSize int64             `form:"pageSize" binding:"required,gte=1"`
	Params   ListInGroupParams `form:"params"`
}

// GetTestFlowRequest
// swagger:model getTestFlowRequest
type GetTestFlowRequest struct {
	ID   *string `form:"id"`
	Name *string `form:"name"`
}

// CountTestFlowRequest
// swagger:model countTestFlowRequest
type CountTestFlowRequest struct {
	GroupID *string `form:"groupId"`
	Name    *string `form:"name"`
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
