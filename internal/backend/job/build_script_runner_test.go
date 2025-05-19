package job

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecScript_ExecScript(t *testing.T) {
	runner := &ExecScript{
		PwdDir:   ".",
		Proxy:    "http://127.0.0.1:7890",
		Registry: "http:/127.0.0.2:1234",
		Env:      map[string]string{},
	}

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		err := runner.ExecScript(ctx, BuildParams{
			Script: "ls",
			Commit: "",
		})
		assert.Nil(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		err := runner.ExecScript(ctx, BuildParams{
			Script: "xx",
			Commit: "",
		})
		assert.Error(t, err)
	})

	t.Run("multipleline", func(t *testing.T) {
		script := `
echo {{.Proxy}}
echo {{.Commit}}
`
		err := runner.ExecScript(ctx, BuildParams{
			Script: script,
			Commit: "aaaa",
		})
		assert.Nil(t, err)
	})

	t.Run("multi runner", func(t *testing.T) {
		script := `
echo {{.Proxy}}
exit(1)
`
		for i := 0; i < 5; i++ {
			err := runner.ExecScript(ctx, BuildParams{
				Script: script,
				Commit: "aaaa",
			})
			assert.NotNil(t, err)
		}
	})
}
