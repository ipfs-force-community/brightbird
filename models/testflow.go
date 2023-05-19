package models

import (
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestFlow
// swagger:model testFlow
type TestFlow struct {
	ID primitive.ObjectID `bson:"_id" json:"id"`
	// the name for this test flow
	// required: true
	// min length: 3
	Name        string              `json:"name"`
	GroupId     primitive.ObjectID  `json:"groupId" bson:"groupid"` //provent mongo use Id to id
	Nodes       []*types.DeployNode `json:"nodes"`
	Cases       []*types.TestItem   `json:"cases"`
	Graph       string              `json:"graph"`
	Description string              `json:"description"`

	BaseTime `bson:",inline"`
}
