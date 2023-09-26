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

// fatal error: stack overflow
type PluginInfo struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	PluginType  PluginType `json:"pluginType"`
	Description string     `json:"description"`

	DeployPluginParams `bson:",inline"`
	PluginParams       `bson:",inline"`
}

type DeployPluginParams struct {
	Repo        string `json:"repo"`
	ImageTarget string `json:"imageTarget"` //use to check image already exit
	BuildScript string `json:"buildScript"`
}

func (dpp DeployPluginParams) Buildable() bool {
	return len(dpp.Repo) > 0 && len(dpp.BuildScript) > 0
}

type PluginParams struct {
	// swagger:ignore
	// 由于Schema定义复杂， go-swagger无法处理循环递归类型的缘故暂时ignore这个类型，之后如果go-swagger能够应对这个问题或者有一个更好的Schema定义的时候可以考虑去掉这个注释
	InputSchema Schema `json:"inputSchema"`
	// swagger:ignore
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
		return bson.TypeBoolean, nil, err
	}

	var doc interface{}
	err = bson.UnmarshalExtJSON(data, false, &doc)
	if err != nil {
		return bson.TypeBoolean, nil, err
	}

	bsonBytes, err := bson.Marshal(doc)
	if err != nil {
		return bson.TypeBoolean, nil, err
	}

	return bson.TypeEmbeddedDocument, bsoncore.AppendDocument(nil, bsonBytes), err
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
