package env

import (
	"context"

	"github.com/hunjixin/brightbird/types"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type IDeployer interface {
	Name() string
	Pods(context.Context) ([]corev1.Pod, error)
	StatefulSet(context.Context) (*appv1.StatefulSet, error)
	Svc(context.Context) (*corev1.Service, error)
	SvcEndpoint() types.Endpoint
	Deploy(context.Context) (err error)

	GetConfig(ctx context.Context) (interface{}, error)
	Update(ctx context.Context, updateCfg interface{}) error
	Params(string) (interface{}, error)
}

type IExec interface {
	Name() string
	Params(string) (interface{}, error)
}

//// The following types are used for components without configuration files or implemation with other lanaguage

// ChainCoUpdate
type ChainCoConfig struct {
	Nodes     []string
	AuthUrl   string
	AuthToken string
}

type VenusWorkerConfig string //just mock here
