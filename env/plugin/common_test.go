package plugin

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPluginInfo(t *testing.T) {
	file := "../../dist/plugins/deploy/chain-co"
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
		cmd, state, isCmd := ReadCMD(CMDSTATEPREFIX + COMPLETELOG)
		assert.True(t, isCmd)
		assert.Equal(t, CMDSTATEPREFIX, cmd)
		assert.Equal(t, COMPLETELOG, state)
	}

	{
		cmd, val, isCmd := ReadCMD(CMDVALPREFIX + "vvvvvvvvvvvvv")
		assert.True(t, isCmd)
		assert.Equal(t, CMDVALPREFIX, cmd)
		assert.Equal(t, "vvvvvvvvvvvvv", val)
	}
}
