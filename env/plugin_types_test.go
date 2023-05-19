package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalJson(t *testing.T) {
	{
		str := `"aaaaaaa"`
		val, err := UnmarshalJson[string]([]byte(str))
		assert.Nil(t, err)
		assert.Equal(t, "aaaaaaa", val)
	}
	{
		str := `10`
		val, err := UnmarshalJson[int]([]byte(str))
		assert.Nil(t, err)
		assert.Equal(t, 10, val)
	}
	{
		type A struct {
			M string
		}
		str := `{"M":"aaaa"}`
		val, err := UnmarshalJson[A]([]byte(str))
		assert.Nil(t, err)
		assert.Equal(t, "aaaa", val.M)
	}
}
