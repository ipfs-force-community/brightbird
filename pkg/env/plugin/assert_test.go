package plugin

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_reportT_Errorf(t *testing.T) {
	assert.True(t, Assert.NotNil(fmt.Errorf("xxaa")))

	t.Run("panic", func(t *testing.T) {
		defer func() {
			assert.NotNil(t, recover())
		}()
		assert.False(t, Assert.Nil(fmt.Errorf("xxaa")))
	})
}
