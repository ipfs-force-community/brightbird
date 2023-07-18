package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/go-address"
	"github.com/swaggest/jsonschema-go"
	"go.mongodb.org/mongo-driver/bson"
)

type A struct {
	ID int `json:"id"`
}

type MyStruct struct {
	Addr address.Address `json:"-"`

	Addr2           address.Address        `json:"addr2"`
	MapT            map[string]interface{} `json:"mapt"`
	BoolVal         bool                   `json:"boolVal"`
	A               A                      `json:"a"`
	ASlice          []A                    `json:"aSlice"`
	AArrary         [10]A                  `json:"aArray"`
	TwoDimissionArr [10][10]string         `json:"twoDimissionArr"`

	NumberArr   [10]string `json:"numArr"`
	NumberSlice []string   `json:"numSlice"`
	StringSlice []string   `json:"strSlice"`
	StringArray [10]string `json:"strArr"`
	APtr        *A         `json:"aPtr"`
	Amount      float64    `json:"amount" minimum:"10.5" example:"20.6" required:"true"`
	Abc         string     `json:"abc" pattern:"[abc]"`
	_           struct{}   `additionalProperties:"false"`                   // Tags of unnamed field are applied to parent schema.
	_           struct{}   `title:"My Struct" description:"Holds my data."` // Multiple unnamed fields can be used.
}

func TestSchema_MarshalUnMarshalBSONValue(t *testing.T) {
	jsonschema.PropertyNameTag("jsonschema")
	reflector := jsonschema.Reflector{}
	defs := map[string]jsonschema.Schema{}
	reflector.DefaultOptions = append(reflector.DefaultOptions,
		jsonschema.DefinitionsPrefix("#/$defs/"),
		jsonschema.CollectDefinitions(func(name string, schema jsonschema.Schema) {
			defs[name] = schema
		}),
		jsonschema.InterceptSchema(func(params jsonschema.InterceptSchemaParams) (stop bool, err error) {
			if params.Value.Type() == reflect.TypeOf(address.Undef) {
				s := jsonschema.Schema{}
				s.AddType(jsonschema.String)
				s.WithExtraPropertiesItem("configurable", true)
				typeName := "FilAddress"
				defs[typeName] = s

				// Replacing current schema with reference.
				rs := jsonschema.Schema{}
				rs.WithRef(fmt.Sprintf("#/$defs/%s", typeName))
				*params.Schema = rs
				params.Processed = true
				return true, nil
			}
			return false, nil
		}))

	schema, err := reflector.Reflect(MyStruct{})
	assert.Nil(t, err)

	schema.WithExtraPropertiesItem("$defs", defs)

	type SchemaWrapper struct {
		Sch *Schema
	}
	sch := (Schema)(schema)
	j, err := bson.Marshal(SchemaWrapper{
		Sch: &sch,
	})
	assert.Nil(t, err)

	var wraper = SchemaWrapper{}
	err = bson.Unmarshal(j, &wraper)
	assert.Nil(t, err)

	jsonBytes1, err := json.Marshal(sch)
	assert.Nil(t, err)

	jsonBytes2, err := json.Marshal(wraper.Sch)
	assert.Nil(t, err)

	assert.Equal(t, len(jsonBytes1), len(jsonBytes2))
}
