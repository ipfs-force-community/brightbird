package env

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParams_Params(t *testing.T) {
	{
		params := ParamsFromVal("123")
		jsonBytes, err := json.Marshal(params)
		assert.Nil(t, err)
		actualParams := Params{}
		err = json.Unmarshal(jsonBytes, &actualParams)
		assert.Nil(t, err)
		actualStr, err := UnmarshalJson[string](actualParams.Raw())
		assert.Nil(t, err)
		assert.Equal(t, "123", actualStr)
	}
	{
		type A struct {
			M string
		}
		params := ParamsFromVal(A{M: "aaaa"})
		jsonBytes, err := json.Marshal(params)
		assert.Nil(t, err)
		actualParams := Params{}
		err = json.Unmarshal(jsonBytes, &actualParams)
		assert.Nil(t, err)
		actual, err := UnmarshalJson[A](actualParams.Raw())
		assert.Nil(t, err)
		assert.Equal(t, "aaaa", actual.M)
	}
	{
		params := ParamsFromVal(10)
		jsonBytes, err := json.Marshal(params)
		assert.Nil(t, err)
		actualParams := Params{}
		err = json.Unmarshal(jsonBytes, &actualParams)
		assert.Nil(t, err)
		actualStr, err := UnmarshalJson[int](actualParams.Raw())
		assert.Nil(t, err)
		assert.Equal(t, 10, actualStr)
	}
}
