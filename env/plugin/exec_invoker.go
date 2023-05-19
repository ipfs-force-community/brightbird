package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/hunjixin/brightbird/env"
)

var _ env.IExec = (*ExecInvoker)(nil)

type ExecInvoker struct {
	client   *http.Client
	sockPath string
}

func NewExecInvoker(sockPath string) (*ExecInvoker, error) {
	httpc := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", sockPath)
			},
		},
	}

	return &ExecInvoker{
		client:   httpc,
		sockPath: sockPath,
	}, nil
}

func (serve *ExecInvoker) get(path string, item interface{}) error {
	resp, err := serve.client.Get("http://unix/" + path)
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
	if item != nil {
		data, err := io.ReadAll(resp.Body)
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

func (serve *ExecInvoker) post(path string, body io.Reader) error {
	resp, err := serve.client.Post("http://unix/"+path, "application/json", nil)
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
	return nil
}

func (serve *ExecInvoker) Param(key string) (env.Params, error) {
	p := env.Params{}
	err := serve.get("params/"+key, &p)
	if err != nil {
		return env.Params{}, err
	}
	return p, nil
}
