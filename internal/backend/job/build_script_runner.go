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

	jobLog.Debugf("exec builb script [%s] : %s", runner.PwdDir, renderResult.String())
	cmd := exec.Command("/bin/sh", "-c", renderResult.String())
	cmd.Dir = runner.PwdDir
	cmd.Env = os.Environ()
	for k, v := range runner.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%v", k, v))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("exec build script [%s] fail : %w | %v", renderResult.String(), err, string(output))
	}

	if !cmd.ProcessState.Success() {
		return fmt.Errorf("exit code not zero %d", cmd.ProcessState.ExitCode())
	}
	return nil
}
