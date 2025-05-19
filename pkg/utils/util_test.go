package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasDupItemInArrary(t *testing.T) {
	assert.False(t, HasDupItemInArrary([]string{"a", "b", "c"}))
	assert.True(t, HasDupItemInArrary([]string{"a", "b", "a"}))
	assert.True(t, HasDupItemInArrary([]string{"a", "a"}))
}
