package utils

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMergeMapStructAndJson(t *testing.T) {
	type A struct {
		P1 map[string]string
	}
	result, err := MergeStructAndJson(A{}, A{}, json.RawMessage(`{"P1":{"Out":"xxx"}}`))
	assert.NoError(t, err)
	assert.Equal(t, result.P1["Out"], "xxx")
}

func TestMergeStructAndJson(t *testing.T) {
	type A struct {
		P1 string
		P2 int
		P3 int
		P4 int
	}

	result, err := MergeStructAndJson(A{
		P1: "aaa",
		P2: 1,
	}, A{
		P1: "bbb",
		P2: 2,
		P3: 1,
	}, json.RawMessage("{}"))
	assert.NoError(t, err)
	assert.Equal(t, result.P1, "bbb")
	assert.Equal(t, result.P2, 2)
	assert.Equal(t, result.P3, 1)

	result, err = MergeStructAndJson(A{
		P1: "aaa",
		P2: 1,
	}, A{
		P1: "bbb",
		P2: 2,
		P3: 1,
	}, json.RawMessage(`{
	"P1":"ccc",
	"P2":4,
	"P3":5,
	"P4":6
}`))
	assert.NoError(t, err)
	assert.Equal(t, result.P1, "ccc")
	assert.Equal(t, result.P2, 4)
	assert.Equal(t, result.P3, 5)
	assert.Equal(t, result.P4, 6)
}
