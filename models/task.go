package models

import (
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// State  task state
// swagger:alias
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
	// Init init state
	Init
	// Running task was running
	Running
	// TempError task got temp error and try late
	TempError
	// Error task  was error and never retry
	Error
	// Successful task success
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
	TestId          types.TestId       `json:"testId"`
	State           State              `json:"state"`
	Logs            []string           `json:"logs"`
	InheritVersions map[string]string  `json:"inheritVersions"` // save a copy of task flow, but task flow update version information in this running
	CommitMap       map[string]string  `json:"versions"`        // save a copy of task flow, but task flow update version information in this running
	BaseTime        `bson:",inline"`
}

func (task Task) InRunning() bool {
	return task.State == Running || task.State == TempError
}
