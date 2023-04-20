package env

import (
	"bytes"
	"context"
	"io"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type env/exec_cmd.goDeployer struct {
	k8sClient *kubernetes.Clientset
	k8sCfg    *rest.Config
	namespace string
}

func (env *Deployer) ExecCmd(ctx context.Context, podName string, cli []string) ([]byte, error) {
	cmd := cli
	req := env.k8sClient.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace(env.namespace).SubResource("exec")
	option := &corev1.PodExecOptions{
		Command: cmd,
		Stdin:   false,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(env.k8sCfg, "POST", req.URL())
	if err != nil {
		return nil, err
	}
	stdOut := bytes.NewBuffer(nil)
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: stdOut,
		Stderr: os.Stderr,
		Tty:    true,
	})
	if err != nil {
		return nil, err
	}
	return io.ReadAll(stdOut)
}
