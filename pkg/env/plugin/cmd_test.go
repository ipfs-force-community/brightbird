package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLastJson(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		result := GetLastJSON(`{"a":123}`)
		assert.Equal(t, `{"a":123}`, result)
	})
	t.Run("multiline", func(t *testing.T) {
		result := GetLastJSON(`aaaaa\nccasd\n{"a":123}`)
		assert.Equal(t, `{"a":123}`, result)
	})
	t.Run("multilineJson", func(t *testing.T) {
		result := GetLastJSON(`aaaaa
		ccasd
		{
			"a"
			:
			123
			}
			`)
		assert.Equal(t, `{
			"a"
			:
			123
			}`, result)
	})
}
