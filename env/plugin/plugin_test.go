package plugin

import (
	"testing"

	"github.com/hunjixin/brightbird/types"
	"github.com/stretchr/testify/assert"
)

func TestConvertValue(t *testing.T) {
	{
		val, err := ConvertValue[string]("aaaa")
		assert.Nil(t, err)
		assert.Equal(t, "aaaa", val)
	}
	{
		val, err := ConvertValue[int8]("127")
		assert.Nil(t, err)
		assert.Equal(t, int8(127), val)
	}
	{
		_, err := ConvertValue[int8]("128")
		assert.NotNil(t, err)
	}
	{
		val, err := ConvertValue[int]("1234")
		assert.Nil(t, err)
		assert.Equal(t, int(1234), val)
	}
	{
		val, err := ConvertValue[bool]("true")
		assert.Nil(t, err)
		assert.Equal(t, true, val)
	}
	{
		val, err := ConvertValue[uint64]("18446744073709551615")
		assert.Nil(t, err)
		assert.Equal(t, uint64(18446744073709551615), val)
	}
}

func TestGetPropertyValue(t *testing.T) {
	{
		val, err := GetPropertyValue(&types.Property{
			Type:  "string",
			Value: "string val",
		})
		assert.Nil(t, err)
		assert.Equal(t, "string val", val)
	}
	{
		val, err := GetPropertyValue(&types.Property{
			Type:  "number",
			Value: "123123123123",
		})
		assert.Nil(t, err)
		assert.Equal(t, int64(123123123123), val)
	}
	{
		val, err := GetPropertyValue(&types.Property{
			Type:  "decimical",
			Value: "12312.3123123",
		})
		assert.Nil(t, err)
		assert.Equal(t, float64(12312.3123123), val)
	}
}
