package env

import (
	"context"
	"errors"
	"reflect"

	"github.com/hunjixin/brightbird/types"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var IDeployerT = reflect.TypeOf((*IDeployer)(nil)).Elem()
var IExecT = reflect.TypeOf((*IExec)(nil)).Elem()

type IDeployer interface {
	Name() string
	Pods(context.Context) ([]corev1.Pod, error)
	StatefulSet(context.Context) (*appv1.StatefulSet, error)
	Svc(context.Context) (*corev1.Service, error)
	SvcEndpoint() types.Endpoint
	Deploy(context.Context) (err error)

	GetConfig(ctx context.Context) (interface{}, error)
	Update(ctx context.Context, updateCfg interface{}) error
	Param(string) (interface{}, error)
}

type IExec interface {
	Param(string) (interface{}, error)
}

var ErrParamsNotFound = errors.New("not found")

type SimpleExec map[string]interface{}

func NewSimpleExec() *SimpleExec {
	return (*SimpleExec)(&map[string]interface{}{})
}

func (exec SimpleExec) Add(key string, val interface{}) SimpleExec {
	exec[key] = val
	return exec
}

func (exec SimpleExec) Param(key string) (interface{}, error) {
	val, ok := exec[key]
	if !ok {
		return nil, ErrParamsNotFound
	}
	return val, nil
}

//// The following types are used for components without configuration files or implemation with other lanaguage

// ChainCoUpdate
type ChainCoConfig struct {
	Nodes     []string
	AuthUrl   string
	AuthToken string
}

type VenusWorkerConfig string //just mock here
