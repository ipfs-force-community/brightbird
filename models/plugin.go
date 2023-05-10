package models

import (
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PluginOut
// swagger:model pluginOut
type PluginOut struct {
	ID       primitive.ObjectID `bson:"_id" json:"id"`
	BaseTime `bson:",inline"`

	types.PluginInfo `bson:",inline"`
	Path             string                   `json:"path"`
	Instance         types.DependencyProperty `json:"instance"`
}
