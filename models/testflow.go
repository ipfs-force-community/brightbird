package models

import (
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/yaml.v3"
)

// TestFlow
// swagger:model testFlow
type TestFlow struct {
	ID primitive.ObjectID `bson:"_id" json:"id"`
	// the name for this test flow
	// required: true
	// min length: 3
	Name        string             `json:"name"`
	GroupId     primitive.ObjectID `json:"groupId" bson:"groupid"` //provent mongo use Id to id
	Graph       string             `json:"graph"`
	Description string             `json:"description"`

	BaseTime `bson:",inline"`
}

type Graph struct {
	Name     string    `yaml:"name"`
	Pipeline Pipelines `yaml:"pipeline"`
	Rawdata  string    `yaml:"raw-data"`
}

// MapSlice encodes and decodes as a YAML map.
// The order of keys is preserved when encoding and decoding.
type Pipelines []PipelineItem

func (h *Pipelines) UnmarshalYAML(value *yaml.Node) error {
	for i := 0; i < len(value.Content); i += 2 {
		var pipline PipelineItem
		if err := value.Content[i+1].Decode(&pipline.Value); err != nil {
			return err
		}
		pipline.Key = value.Content[i].Value

		*h = append(*h, pipline)
	}

	return nil
}

// MapItem is an item in a MapSlice.
type PipelineItem struct {
	Key   string
	Value *types.ExecNode
}
