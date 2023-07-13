package types

import (
	"errors"

	"github.com/swaggest/jsonschema-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// PluginType  type of plugin
// swagger:alias
type PluginType string

const (
	// Deploy deploy conponet
	Deploy PluginType = "Deployer"
	// TestExec test case
	TestExec PluginType = "Exec"
)

type PluginInfo struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	PluginType  PluginType `json:"pluginType"`
	Description string     `json:"description"`
	Repo        string     `json:"repo"`
	ImageTarget string     `json:"imageTarget"`

	PluginParams `bson:",inline"`
}

type PluginParams struct {
	InputSchema  Schema `json:"inputSchema"`
	OutputSchema Schema `json:"outputSchema"`
}

type Schema jsonschema.Schema

func (schema Schema) MarshalJSON() ([]byte, error) {
	return (jsonschema.Schema)(schema).MarshalJSON()
}

func (schema *Schema) UnmarshalJSON(data []byte) error {
	return (*jsonschema.Schema)(schema).UnmarshalJSON(data)
}

func (schema Schema) MarshalBSONValue() (bsontype.Type, []byte, error) {
	data, err := (jsonschema.Schema)(schema).MarshalJSON()
	if err != nil {
		return bsontype.Boolean, nil, err
	}

	var doc interface{}
	err = bson.UnmarshalExtJSON(data, false, &doc)
	if err != nil {
		return bsontype.Boolean, nil, err
	}

	bsonBytes, err := bson.Marshal(doc)
	if err != nil {
		return bsontype.Boolean, nil, err
	}

	return bsontype.EmbeddedDocument, bsoncore.AppendDocument(nil, bsonBytes), err
}

func (schema *Schema) UnmarshalBSONValue(t bsontype.Type, b []byte) error {
	bsonDoc, _, ok := bsoncore.ReadDocument(b)
	if !ok {
		return errors.New("Schema UnmarshalBSONValue error")
	}

	var doc interface{}
	err := bson.Unmarshal(bsonDoc, &doc)
	if err != nil {
		return err
	}

	jsonBytes, err := bson.MarshalExtJSON(doc, false, false)
	if err != nil {
		return err
	}

	jsonSchema := &jsonschema.Schema{}
	err = jsonSchema.UnmarshalJSON(jsonBytes)
	if err != nil {
		return err
	}
	*schema = Schema(*jsonSchema)
	return nil
}
