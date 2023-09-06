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

type ExecScript struct {
	GitToken string
	PwdDir   string
	Proxy    string
	Registry string
	Env      map[string]string
}

// plugin + version
func (runner *ExecScript) ExecScript(ctx context.Context, params BuildParams) error {
	//render template
	t, err := template.New("").Parse(params.Script)
	if err != nil {
		return err
	}

	renderResult := bytes.NewBuffer(nil)
	err = t.Execute(renderResult, struct {
		GitToken string
		Commit   string
		Proxy    string
		Registry string
	}{
		runner.GitToken,
		params.Commit,
		runner.Proxy,
		runner.Registry,
	})
	if err != nil {
		return err
	}

	log.Debugf("space %s script %s", runner.PwdDir, renderResult.String())
	cmd := exec.Command("/bin/sh", "-c", renderResult.String())
	cmd.Dir = runner.PwdDir
	cmd.Env = os.Environ()
	for k, v := range runner.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%v", k, v))
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("exec build script %s fail %w", renderResult.String(), err)
	}

	if !cmd.ProcessState.Success() {
		return fmt.Errorf("exit code not zero %d", cmd.ProcessState.ExitCode())
	}
	return nil
}
