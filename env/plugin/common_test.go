package plugin

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPluginInfo(t *testing.T) {
	file := "../../dist/plugins/deploy/sophon-co"
	info, err := GetPluginInfo(file)
	assert.NoError(t, err)
	reg, err := regexp.Compile(`^v[\d]+?.[\d]+?.[\d]+?$`)
	assert.NoError(t, err)
	assert.True(t, reg.MatchString(info.Version))
}

func TestReadCMD(t *testing.T) {
	{
		cmd, errMsg, isCmd := ReadCMD(CMDERRORREFIX + "error")
		assert.True(t, isCmd)
		assert.Equal(t, CMDERRORREFIX, cmd)
		assert.Equal(t, "error", errMsg)
	}

	{
		cmd, addition, isCmd := ReadCMD(CMDSTARTPREFIX + "test")
		assert.True(t, isCmd)
		assert.Equal(t, CMDSTARTPREFIX, cmd)
		assert.Equal(t, "test", addition)
	}

	{
		cmd, addition, isCmd := ReadCMD(CMDSUCCESSPREFIX + "test")
		assert.True(t, isCmd)
		assert.Equal(t, CMDSUCCESSPREFIX, cmd)
		assert.Equal(t, "test", addition)
	}

	{
		cmd, val, isCmd := ReadCMD(CMDVALPREFIX + "vvvvvvvvvvvvv")
		assert.True(t, isCmd)
		assert.Equal(t, CMDVALPREFIX, cmd)
		assert.Equal(t, "vvvvvvvvvvvvv", val)
	}
}
