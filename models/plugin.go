package models

import (
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PluginDetail
// swagger:model pluginDetail
type PluginDetail struct {
	ID       primitive.ObjectID `bson:"_id" json:"id"`
	BaseTime `bson:",inline"`

	types.PluginInfo `bson:",inline"`
	Path             string                   `json:"path"`
	Instance         types.DependencyProperty `json:"instance"`
}

// PluginInfo
// swagger:model pluginInfo
type PluginInfo struct {
	Name        string           `json:"name"`
	PluginType  types.PluginType `json:"pluginType"`
	Description string           `json:"description"`
}
