package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsAnnotateOut(t *testing.T) {
	type A struct {
		AnnotateOut
	}
	assert.True(t, IsAnnotateOut(A{}))
	assert.False(t, IsAnnotateOut(&A{}))

}
