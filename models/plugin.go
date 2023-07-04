package models

import (
	"fmt"

	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Plugin
// swagger:model plugin
type PluginDef struct {
	types.PluginInfo `bson:",inline"`
	Path             string `json:"path"`
}

func (p PluginDef) GetInputProperty(byName string) (*types.Property, error) {
	for _, prop := range p.InputProperties {
		if prop.Name == byName {
			return &prop, nil
		}
	}
	return nil, fmt.Errorf("property %s not found", byName)
}

func (p PluginDef) GetOutputProperty(byName string) (*types.Property, error) {
	for _, prop := range p.OutputProperties {
		if prop.Name == byName {
			return &prop, nil
		}
	}
	return nil, fmt.Errorf("property %s not found", byName)
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
