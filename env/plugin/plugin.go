package plugin

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/filecoin-project/go-address"
	"github.com/swaggest/jsonschema-go"
)

func ParserSchema(t reflect.Type) (jsonschema.Schema, error) {
	reflector := jsonschema.Reflector{}
	defs := map[string]jsonschema.SchemaOrBool{}
	reflector.DefaultOptions = append(reflector.DefaultOptions,
		jsonschema.ProcessWithoutTags,
		jsonschema.PropertyNameTag("jsonschema"),
		jsonschema.CollectDefinitions(func(name string, schema jsonschema.Schema) {
			defs[name] = schema.ToSchemaOrBool()
		}),
		jsonschema.InterceptSchema(func(params jsonschema.InterceptSchemaParams) (stop bool, err error) {
			if params.Value.Type() == reflect.TypeOf(address.Undef) {
				s := jsonschema.Schema{}
				s.AddType(jsonschema.String)
				s.WithExtraPropertiesItem("configurable", true)
				typeName := "FilAddress"
				defs[typeName] = s.ToSchemaOrBool()

				// Replacing current schema with reference.
				rs := jsonschema.Schema{}
				rs.WithRef(fmt.Sprintf("#/definitions/%s", typeName))
				*params.Schema = rs
				params.Processed = true
				return true, nil
			}
			return false, nil
		}))

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	val := reflect.Indirect(reflect.New(t).Elem()).Interface()
	schema, err := reflector.Reflect(val)
	if err != nil {
		return jsonschema.Schema{}, err
	}
	schema.WithDefinitions(defs)
	return schema, nil
}

// path a
// path a.b
// path a[0].b   object
// path a[0][0]  matrics
// path a[0]     simple type

type SchemaPropertyFinder struct {
	gSchema jsonschema.Schema
}

func NewSchemaPropertyFinder(gSchema jsonschema.Schema) *SchemaPropertyFinder {
	return &SchemaPropertyFinder{gSchema: gSchema}
}

func (finder *SchemaPropertyFinder) resolveChildType(definitions map[string]jsonschema.SchemaOrBool, schema *jsonschema.Schema) *jsonschema.Schema {
	if schema.Ref != nil {
		refKey := strings.ReplaceAll(*schema.Ref, "#/definitions/", "")
		def, ok := definitions[refKey]
		if ok {
			return def.TypeObject
		}
		panic("not found")
	}

	return schema
}

func (finder *SchemaPropertyFinder) FindPath(path string) (jsonschema.SimpleType, error) {
	defs := finder.gSchema.Definitions
	pathSeq, err := SplitJsonPath(path)
	if err != nil {
		return jsonschema.Null, err
	}
	currentSchema := &finder.gSchema
	for i := 0; i < len(pathSeq); i++ {
		seq := pathSeq[i]
		if seq.IsFirst {
			//specific case first
			val, ok := currentSchema.Properties[seq.Name]
			if !ok {
				return jsonschema.Null, fmt.Errorf("property %v not found", seq)
			}

			currentSchema = finder.resolveChildType(defs, val.TypeObject)
			if seq.IsLast {
				return getSchemaType(currentSchema), nil
			}
			continue
		}

		if currentSchema.Type.SliceOfSimpleTypeValues != nil && currentSchema.Items != nil {
			//array
			if !seq.IsIndex {
				return jsonschema.Null, fmt.Errorf("schema is array but path not %v", seq)
			}
			if seq.IsLast {
				return getSchemaType(currentSchema.Items.SchemaOrBool.TypeObject), nil
			}
			currentSchema = finder.resolveChildType(defs, currentSchema.Items.SchemaOrBool.TypeObject)
			continue
		}

		switch *currentSchema.Type.SimpleTypes {
		case jsonschema.Object:
			//specific case first
			val, ok := currentSchema.Properties[seq.Name]
			if !ok {
				return jsonschema.Null, fmt.Errorf("property %v not found", seq)
			}
			currentSchema = finder.resolveChildType(defs, val.TypeObject)
			if seq.IsLast {
				return getSchemaType(currentSchema), nil
			}
			continue
		case jsonschema.String:
			fallthrough
		case jsonschema.Boolean:
			fallthrough
		case jsonschema.Integer:
			fallthrough
		case jsonschema.Null:
			fallthrough
		case jsonschema.Number:
			return *currentSchema.Type.SimpleTypes, nil
		default:
			return jsonschema.Null, fmt.Errorf("type should be chekc before")
		}
	}

	return jsonschema.Null, fmt.Errorf("path and schema not match , path have more path than schema %s", path)
}

func getSchemaType(schema *jsonschema.Schema) jsonschema.SimpleType {
	if schema.Type.SliceOfSimpleTypeValues != nil && schema.Items != nil {
		//array
		return jsonschema.Array
	}

	return *schema.Type.SimpleTypes
}

type JsonPathSec struct {
	Index   int
	IsIndex bool
	IsArray bool
	Name    string
	IsLast  bool
	IsFirst bool
}

func SplitJsonPath(path string) ([]JsonPathSec, error) {
	var result []JsonPathSec
	for _, seq := range strings.Split(path, ".") {
		index, err := strconv.Atoi(seq)
		if err != nil {
			//not index
			result = append(result, JsonPathSec{
				Name: seq,
			})
			continue
		}
		//change pre to array
		result[len(result)-1].IsArray = true
		result = append(result, JsonPathSec{
			Name:    "[]",
			IsIndex: true,
			Index:   index,
		})
	}

	if len(result) > 0 {
		result[0].IsFirst = true
		result[len(result)-1].IsLast = true
	}
	return result, nil
}

func GetJsonValue(schemaType jsonschema.SimpleType, value string) (interface{}, error) {
	switch schemaType {
	case jsonschema.Boolean:
		return value == "true", nil
	case jsonschema.Integer:
		intVal, ok := big.NewInt(0).SetString(value, 10)
		if !ok {
			return nil, fmt.Errorf("parser json number(%s) failed", value)
		}
		return intVal.Int64(), nil
	case jsonschema.Number: //todo consider big number
		if strings.Contains(value, ".") {
			rat := &big.Rat{}
			rat, ok := rat.SetString(value)
			if !ok {
				return nil, fmt.Errorf("parser json number(%s) failed", value)
			}
			float64Val, _ := rat.Float64()
			return float64Val, nil
		}
		intVal, ok := big.NewInt(0).SetString(value, 10)
		if !ok {
			return nil, fmt.Errorf("parser json number(%s) failed", value)
		}
		return intVal.Int64(), nil
	case jsonschema.String:
		return value, nil
	case jsonschema.Object:
		var jsonRaw interface{}
		err := json.Unmarshal([]byte(value), &jsonRaw)
		if err != nil {
			return nil, err
		}
		return jsonRaw, nil
	case jsonschema.Array:
		var jsonRaw []interface{}
		err := json.Unmarshal([]byte(value), &jsonRaw)
		if err != nil {
			return nil, err
		}
		return jsonRaw, nil
	}
	return nil, fmt.Errorf("unsupport property type %s", schemaType)
}
