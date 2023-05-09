package types

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PageReq[T any] struct {
	PageNum  int64 `form:"pageNum" binding:"required,gte=1"`
	PageSize int64 `form:"pageSize" binding:"required,gte=1"`
	Params   T     `form:"params"`
}

func (pageReq PageReq[T]) Skip() int64 {
	if pageReq.PageNum < 1 {
		return 0
	}
	return (pageReq.PageNum - 1) * pageReq.PageSize
}

func (pageReq PageReq[T]) Take() int64 {
	return pageReq.PageSize
}

type PageResp[T any] struct {
	Total   int64 `json:"total"`
	Pages   int64 `json:"pages"`
	PageNum int64 `json:"pageNum"`
	List    []T   `json:"list"`
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
	ID            primitive.ObjectID `bson:"_id" json:"id"`
	BaseTime      `bson:",inline"`
	PluginInfo    `bson:",inline"`
	Properties    []Property `json:"properties"`
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

// SharedPropertyInNode just use to get shared field between deploynode and testitem, no pratical usage
type SharedPropertyInNode interface {
	GetName() string
	GetProperties() []*Property
	GetSvcProperties() []*Property
	GetOut() *Property
}

type DeployNode struct {
	// the name for this test flow
	// required: true
	// min length: 3
	Name          string      `json:"name"`
	Properties    []*Property `json:"properties"`
	SvcProperties []*Property `json:"svcProperties"`
	Out           *Property   `json:"out"`
}

func (n DeployNode) GetName() string               { return n.Name }
func (n DeployNode) GetProperties() []*Property    { return n.Properties }
func (n DeployNode) GetSvcProperties() []*Property { return n.SvcProperties }
func (n DeployNode) GetOut() *Property             { return n.Out }

type TestItem struct {
	// the name for this test flow
	// required: true
	// min length: 3
	Name          string      `json:"name"`
	Properties    []*Property `json:"properties"`
	SvcProperties []*Property `json:"svcProperties"`
	Out           *Property   `json:"out"`
}

func (n TestItem) GetName() string               { return n.Name }
func (n TestItem) GetProperties() []*Property    { return n.Properties }
func (n TestItem) GetSvcProperties() []*Property { return n.SvcProperties }
func (n TestItem) GetOut() *Property             { return n.Out }

// TestFlow
// swagger:model testFlow
type TestFlow struct {
	ID primitive.ObjectID `bson:"_id" json:"id"`
	// the name for this test flow
	// required: true
	// min length: 3
	Name        string             `json:"name"`
	GroupId     primitive.ObjectID `json:"groupId" bson:"groupid"` //provent mongo use Id to id
	Nodes       []*DeployNode      `json:"nodes"`
	Cases       []*TestItem        `json:"cases"`
	Graph       string             `json:"graph"`
	Description string             `json:"description"`

	BaseTime `bson:",inline"`
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

	BaseTime `bson:",inline"`
}

type JobType string

const (
	CronJobType       JobType = "cron_job"
	PRMergedJobType   JobType = "pr_merged_hook"
	TagCreatedJobType JobType = "tag_created_hook"
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
	PRMergedJobParams
	TagCreateJobParams

	BaseTime `bson:",inline"`
}

func (job Job) CheckParams() error {
	switch job.JobType {
	case CronJobType:
		_, err := cron.ParseStandard(job.CronExpression)
		return err
	case PRMergedJobType:
		for _, match := range job.PRMergedJobParams.PRMergedEventMatchs {
			if len(match.BasePattern) == 0 || len(match.SourcePattern) == 0 {
				return errors.New("pr merged job must have dest and source branch")
			}
			_, err := regexp.Compile(match.BasePattern)
			if err != nil {
				return fmt.Errorf("%s not a correct regex pattern %v", match.BasePattern, err)
			}

			_, err = regexp.Compile(match.SourcePattern)
			if err != nil {
				return fmt.Errorf("%s not a correct regex pattern %v", match.SourcePattern, err)
			}
		}

	case TagCreatedJobType:
		for _, match := range job.TagCreateEventMatchs {
			if len(match.TagPattern) == 0 {
				return errors.New("tag create event must have a name")
			}

			_, err := regexp.Compile(match.TagPattern)
			if err != nil {
				return fmt.Errorf("%s not a correct regex pattern %v", match.TagPattern, err)
			}
		}
	default:
		return fmt.Errorf("unsupport job type")
	}
	return nil
}

type CronJobParams struct {
	CronExpression string `json:"cronExpression"`
}

type PRMergedJobParams struct {
	PRMergedEventMatchs []PRMergedEventMatch `json:"prMergedEventMatchs"`
}

type PRMergedEventMatch struct {
	Repo          string `json:"repo"`
	BasePattern   string `json:"basePattern"`
	SourcePattern string `json:"sourcePattern"`
}

type TagCreateJobParams struct {
	TagCreateEventMatchs []TagCreateEventMatch `json:"tagCreateEventMatchs"`
}

type TagCreateEventMatch struct {
	Repo       string `json:"repo"`
	TagPattern string `json:"tagPattern"`
}

type State int

func (st State) Stringer() string {
	return st.String()
}

func (st State) String() string {
	switch st {
	case Init:
		return "init"
	case Running:
		return "running"
	case TempError:
		return "temperr"
	case Error:
		return "error"
	case Successful:
		return "success"
	}
	return ""
}

const (
	_ State = iota //skip default 0
	Init
	Running
	TempError
	Error
	Successful
)

// Task
// swagger:model task
type Task struct {
	ID              primitive.ObjectID `bson:"_id" json:"id"`
	Name            string             `json:"name"`
	PodName         string             `json:"podName"`
	JobId           primitive.ObjectID `json:"jobId"`
	TestFlowId      primitive.ObjectID `json:"testFlowId"` //save this field for convience, get from job info is alright
	TestId          TestId             `json:"testId"`
	State           State              `json:"state"`
	Logs            []string           `json:"logs"`
	InheritVersions map[string]string  `json:"inheritVersions"` // save a copy of task flow, but task flow update version information in this running
	CommitMap       map[string]string  `json:"versions"`        // save a copy of task flow, but task flow update version information in this running
	BaseTime        `bson:",inline"`
}

func (task Task) InRunning() bool {
	return task.State == Running || task.State == TempError
}
