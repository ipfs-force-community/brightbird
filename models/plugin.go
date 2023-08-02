package models

import (
	"github.com/ipfs-force-community/brightbird/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Plugin
// swagger:model plugin
type PluginDef struct {
	types.PluginInfo `bson:",inline"`
	Path             string `json:"path"`
}

// PluginDetail
// swagger:model pluginDetail
type PluginDetail struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Name        string             `json:"name"`
	PluginType  types.PluginType   `json:"pluginType"`
	Description string             `json:"description"`
	Labels      []string           `json:"labels"`
	PluginDefs  []PluginDef        `json:"pluginDefs"`
	BaseTime    `bson:",inline"`
}
