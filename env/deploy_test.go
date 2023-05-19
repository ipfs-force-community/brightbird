package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleExec_Param(t *testing.T) {
	exec := NewSimpleExec().Add("t1", ParamsFromVal("b1"))
	val, err := exec.Param("t1")
	assert.Nil(t, err)
	assert.Equal(t, "b1", val.MustString())

	_, err = exec.Param("t2")
	assert.Equal(t, ErrParamsNotFound, err)
}
