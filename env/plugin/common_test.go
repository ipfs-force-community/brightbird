package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPluginInfo(t *testing.T) {
	GetPluginInfo("/root/brightbird/plugins/deploy/chain-co")
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
