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

func (p PluginDef) GetInputProperty(namePath string) (*types.Property, error) {
	var prop, found = findProperty(p.InputProperties, namePath)
	if !found {
		fmt.Errorf("property %s not found", namePath)
	}
	return prop, nil
}

func (p PluginDef) GetOutputProperty(namePath string) (*types.Property, error) {
	var prop, found = findProperty(p.OutputProperties, namePath)
	if !found {
		fmt.Errorf("property %s not found", namePath)
	}
	return prop, nil
}

func findProperty(props []types.Property, namePath string) (*types.Property, bool) {
	for _, prop := range props {
		if prop.NamePath == namePath {
			return &prop, true
		}

		if len(prop.Chindren) > 0 {
			findP, found := findProperty(prop.Chindren, namePath)
			if found {
				return findP, found
			}
		}
	}
	return nil, false
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
