package plugin

import (
	"reflect"
	"testing"

	"github.com/filecoin-project/go-address"

	"github.com/swaggest/jsonschema-go"

	"github.com/stretchr/testify/assert"
)

func TestSplitJsonPath(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		path := SplitJSONPath("p1")
		assert.Len(t, path, 1)
		assert.Equal(t, path[0], JSONPathSec{
			Index:   0,
			IsIndex: false,
			IsArray: false,
			Name:    "p1",
			IsLast:  true,
			IsFirst: true,
		})
	})

	t.Run("test first last", func(t *testing.T) {
		path := SplitJSONPath("p1.p2.p3")
		assert.Len(t, path, 3)
		assert.Equal(t, path[0], JSONPathSec{
			Index:   0,
			IsIndex: false,
			IsArray: false,
			Name:    "p1",
			IsLast:  false,
			IsFirst: true,
		})
		assert.Equal(t, path[1], JSONPathSec{
			Index:   0,
			IsIndex: false,
			IsArray: false,
			Name:    "p2",
			IsLast:  false,
			IsFirst: false,
		})
		assert.Equal(t, path[2], JSONPathSec{
			Index:   0,
			IsIndex: false,
			IsArray: false,
			Name:    "p3",
			IsLast:  true,
			IsFirst: false,
		})
	})

	t.Run("array", func(t *testing.T) {
		path := SplitJSONPath("p1.p2.1.2.p3")
		assert.Len(t, path, 5)
		assert.Equal(t, path[0], JSONPathSec{
			Index:   0,
			IsIndex: false,
			IsArray: false,
			Name:    "p1",
			IsLast:  false,
			IsFirst: true,
		})
		assert.Equal(t, path[1], JSONPathSec{
			Index:   0,
			IsIndex: false,
			IsArray: true,
			Name:    "p2",
			IsLast:  false,
			IsFirst: false,
		})
		assert.Equal(t, path[2], JSONPathSec{
			Index:   1,
			IsIndex: true,
			IsArray: true,
			Name:    "[]",
			IsLast:  false,
			IsFirst: false,
		})
		assert.Equal(t, path[3], JSONPathSec{
			Index:   2,
			IsIndex: true,
			IsArray: false,
			Name:    "[]",
			IsLast:  false,
			IsFirst: false,
		})
		assert.Equal(t, path[4], JSONPathSec{
			Index:   0,
			IsIndex: false,
			IsArray: false,
			Name:    "p3",
			IsLast:  true,
			IsFirst: false,
		})
	})
}

func TestSchemaPropertyFinder_FindPath(t *testing.T) {
	type Embed struct {
		Num       int             `jsonschema:"num"`
		SimpleArr []int           `jsonschema:"simpleArr"`
		Addr      address.Address `jsonschema:"addr"`
	}
	type T struct {
		Num         int     `jsonschema:"num"`
		Str         string  `jsonschema:"str"`
		Float       float64 `jsonschema:"floatv"`
		SimpleArr   []int   `jsonschema:"simpleArr"`
		EmbedArr    []Embed `jsonschema:"embedArr"`
		InnerStruct Embed   `jsonschema:"innerStruct"`

		Ignore int `jsonschema:"-"`

		Addr address.Address `jsonschema:"addr"`
	}

	schema, err := ParserSchema(reflect.TypeOf(T{}))
	assert.Nil(t, err)

	finder := NewSchemaPropertyFinder(schema)

	{
		_, err = finder.FindPath("Ignore")
		assert.Contains(t, err.Error(), "not found")
	}

	{
		jsonType, err := finder.FindPath("num")
		assert.Nil(t, err)
		assert.Equal(t, jsonschema.Integer, jsonType)
	}

	{
		jsonType, err := finder.FindPath("str")
		assert.Nil(t, err)
		assert.Equal(t, jsonschema.String, jsonType)
	}

	{
		jsonType, err := finder.FindPath("floatv")
		assert.Nil(t, err)
		assert.Equal(t, jsonschema.Number, jsonType)
	}

	{
		jsonType, err := finder.FindPath("simpleArr")
		assert.Nil(t, err)
		assert.Equal(t, jsonschema.Array, jsonType)
	}

	{
		jsonType, err := finder.FindPath("simpleArr.0")
		assert.Nil(t, err)
		assert.Equal(t, jsonschema.Integer, jsonType)
	}

	{
		jsonType, err := finder.FindPath("embedArr")
		assert.Nil(t, err)
		assert.Equal(t, jsonschema.Array, jsonType)
	}

	{
		jsonType, err := finder.FindPath("embedArr.0.num")
		assert.Nil(t, err)
		assert.Equal(t, jsonschema.Integer, jsonType)
	}
	{
		jsonType, err := finder.FindPath("embedArr.0.addr")
		assert.Nil(t, err)
		assert.Equal(t, jsonschema.String, jsonType)
	}

	{
		jsonType, err := finder.FindPath("embedArr.0.simpleArr.0")
		assert.Nil(t, err)
		assert.Equal(t, jsonschema.Integer, jsonType)
	}
}

func TestGetJsonValue(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		ret, err := GetJSONValue(jsonschema.String, "aaa")
		assert.Nil(t, err)
		assert.Equal(t, "aaa", ret)
	})

	t.Run("integer", func(t *testing.T) {
		ret, err := GetJSONValue(jsonschema.Integer, "123")
		assert.Nil(t, err)
		assert.Equal(t, int64(123), ret)
	})

	t.Run("string", func(t *testing.T) {
		ret, err := GetJSONValue(jsonschema.Number, "1.2312")
		assert.Nil(t, err)
		assert.Equal(t, float64(1.2312), ret)
	})

	t.Run("bool", func(t *testing.T) {
		ret, err := GetJSONValue(jsonschema.Boolean, "true")
		assert.Nil(t, err)
		assert.Equal(t, true, ret)

		ret, err = GetJSONValue(jsonschema.Boolean, "false")
		assert.Nil(t, err)
		assert.Equal(t, false, ret)
	})

	t.Run("object", func(t *testing.T) {
		ret, err := GetJSONValue(jsonschema.Object, `{"a":1}`)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), ret.(map[string]interface{})["a"].(float64))
	})
	t.Run("array", func(t *testing.T) {
		ret, err := GetJSONValue(jsonschema.Object, `[1,2,3,4]`)
		assert.Nil(t, err)
		assert.Equal(t, float64(2), ret.([]interface{})[1].(float64))
	})
}
