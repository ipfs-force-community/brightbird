package job

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"text/template"
)

type BuildParams struct {
	Script string
	Commit string
}

type BuildScriptRunner struct {
	PwdDir   string
	Proxy    string
	Registry string
	Env      map[string]string
}

// plugin + version
func (runner *BuildScriptRunner) BuildScriptRunner(ctx context.Context, params BuildParams) error {
	//render template
	t, err := template.New("").Parse(params.Script)
	if err != nil {
		return err
	}

	renderResult := bytes.NewBuffer(nil)
	err = t.Execute(renderResult, struct {
		Commit   string
		Proxy    string
		Registry string
	}{
		params.Commit,
		runner.Proxy,
		runner.Registry,
	})
	if err != nil {
		return err
	}

	cmd := exec.Command("/bin/sh", "-c", renderResult.String())
	cmd.Dir = runner.PwdDir
	cmd.Env = os.Environ()
	for k, v := range runner.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%v", k, v))
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
