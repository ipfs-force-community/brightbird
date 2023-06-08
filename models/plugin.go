package models

import (
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Plugin
// swagger:model plugin
type Plugin struct {
	types.PluginInfo `bson:",inline"`
	Path             string                   `json:"path"`
	Instance         types.DependencyProperty `json:"instance"`
}

// PluginDetail
// swagger:model pluginDetail
type PluginDetail struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Name        string             `json:"name"`
	PluginType  types.PluginType   `json:"pluginType"`
	Description string             `json:"description"`
	Labels      []string           `json:"labels"`
	Plugins     []Plugin           `json:"plugins"`
	BaseTime    `bson:",inline"`
}
