package env

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/hunjixin/brightbird/types"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var ErrParamsNotFound = errors.New("not found")

var IDeployerT = reflect.TypeOf((*IDeployer)(nil)).Elem()
var IExecT = reflect.TypeOf((*IExec)(nil)).Elem()

type IDeployer interface {
	InstanceName() (string, error)
	Pods(context.Context) ([]corev1.Pod, error)
	StatefulSet(context.Context) (*appv1.StatefulSet, error)
	Svc(context.Context) (*corev1.Service, error)
	SvcEndpoint() (types.Endpoint, error)
	Deploy(context.Context) error

	GetConfig(ctx context.Context) (Params, error)
	Update(ctx context.Context, updateCfg interface{}) error
	Param(string) (Params, error) //todo change method to Param<T(key string) (T, error)  after golang support method generic, issue: https://github.com/golang/go/issues/49085
}

type IExec interface {
	Param(string) (Params, error)
}

type SimpleExec map[string]Params

func NewSimpleExec() *SimpleExec {
	return (*SimpleExec)(&map[string]Params{})
}

func (exec SimpleExec) Add(key string, val Params) SimpleExec {
	exec[key] = val
	return exec
}

func (exec SimpleExec) Param(key string) (Params, error) {
	val, ok := exec[key]
	if !ok {
		return Params{}, ErrParamsNotFound
	}
	return val, nil
}

type Params struct {
	V []byte
}

func ParamsFromVal(val interface{}) Params {
	data, err := json.Marshal(val)
	if err != nil {
		panic("marshal val fail")
	}
	return Params{
		V: data,
	}
}

func (params Params) Marshal() ([]byte, error) {
	return json.Marshal(params.V)
}

func (params Params) JsonUnmarshal(val interface{}) error {
	return nil
}

func (params Params) String() string {
	var val string
	err := json.Unmarshal(params.V, &val)
	if err != nil {
		panic("marshal val fail")
	}
	return val
}
