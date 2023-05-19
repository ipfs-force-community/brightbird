package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/hunjixin/brightbird/env"

	"github.com/hunjixin/brightbird/types"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ env.IDeployer = (*DeployInvoker)(nil)

type DeployInvoker struct {
	client   *http.Client
	sockPath string
}

func NewDeployInvoker(sockPath string) (*DeployInvoker, error) {
	httpc := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", sockPath)
			},
		},
	}

	return &DeployInvoker{
		client:   httpc,
		sockPath: sockPath,
	}, nil
}

func (serve *DeployInvoker) get(path string, item interface{}) error {
	resp, err := serve.client.Get("http://unix/" + path)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		reason, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("request fail reason: %s", string(reason))
	}
	if item != nil {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (serve *DeployInvoker) post(path string, body io.Reader) error {
	resp, err := serve.client.Post("http://unix/"+path, "application/json", nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		reason, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("request fail reason: %s", string(reason))
	}
	return nil
}

func (serve *DeployInvoker) InstanceName() (string, error) {
	p := ""
	err := serve.get("instancename", &p)
	if err != nil {
		return "", err
	}
	return p, err
}

// Deploy implements IDeployer
func (serve *DeployInvoker) Deploy(context.Context) (err error) {
	return serve.post("deploy", nil)
}

// GetConfig implements IDeployer
func (serve *DeployInvoker) GetConfig(ctx context.Context) (env.Params, error) {
	p := env.Params{}
	err := serve.get("getconfig", &p)
	if err != nil {
		return env.Params{}, err
	}
	return p, nil
}

// Param implements IDeployer
func (serve *DeployInvoker) Param(key string) (env.Params, error) {
	p := env.Params{}
	err := serve.get("params/"+key, &p)
	if err != nil {
		return env.Params{}, err
	}
	return p, nil
}

// Pods implements IDeployer
func (serve *DeployInvoker) Pods(context.Context) ([]corev1.Pod, error) {
	pods := []corev1.Pod{}
	err := serve.get("pods", &pods)
	if err != nil {
		return nil, err
	}
	return pods, nil
}

func (serve *DeployInvoker) Stop(ctx context.Context) error {
	//not support for invoker
	panic("not support ")
}

// StatefulSet implements IDeployer
func (serve *DeployInvoker) StatefulSet(context.Context) (*appv1.StatefulSet, error) {
	statefulSet := appv1.StatefulSet{}
	err := serve.get("statefulset", &statefulSet)
	if err != nil {
		return nil, err
	}
	return &statefulSet, nil
}

// Svc implements IDeployer
func (serve *DeployInvoker) Svc(context.Context) (*corev1.Service, error) {
	svc := corev1.Service{}
	err := serve.get("svc", &svc)
	if err != nil {
		return nil, err
	}
	return &svc, nil
}

// SvcEndpoint implements IDeployer
func (serve *DeployInvoker) SvcEndpoint() (types.Endpoint, error) {
	endpoint := types.Endpoint("")
	err := serve.get("svcendpoint", &endpoint)
	if err != nil {
		return "", err
	}

	return endpoint, nil
}

// Update implements IDeployer
func (serve *DeployInvoker) Update(ctx context.Context, updateCfg interface{}) error {
	data, err := json.Marshal(updateCfg)
	if err != nil {
		return err
	}

	return serve.post("update", bytes.NewReader(data))
}
