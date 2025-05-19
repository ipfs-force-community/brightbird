package env

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBaseConfig(t *testing.T) {
	type A struct {
		BaseConfig
		B string
	}

	v := reflect.ValueOf(&A{})
	field := v.Elem().FieldByName("BaseConfig")
	assert.True(t, field.CanAddr())
	field.Set(reflect.ValueOf(NewBaseConfig("x", "y")))
	val := v.Interface().(*A)
	assert.Equal(t, "x", val.CodeVersion)
	assert.Equal(t, "y", val.InstanceName)
}
