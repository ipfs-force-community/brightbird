package models

import "go.mongodb.org/mongo-driver/bson/primitive"

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
