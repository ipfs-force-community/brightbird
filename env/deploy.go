package env

import (
	"context"

	"github.com/hunjixin/brightbird/types"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type IVenusDeployer IDeployer
type IChainCoDeployer interface {
	IDeployer
}

type IMarketClientDeployer IDeployer
type IVenusAuthDeployer IDeployer
type IVenusGatewayDeployer IDeployer
type IVenusMarketDeployer IDeployer
type IVenusMessageDeployer interface {
	IDeployer
}
type IVenusMinerDeployer IDeployer
type IVenusWalletDeployer IDeployer
type IVenusSectorManagerDeployer IDeployer
type IVenusWorkerDeployer IDeployer
type ITestDeployer IDeployer

type IDeployer interface {
	Name() string
	Pods(context.Context) ([]corev1.Pod, error)
	StatefulSet(context.Context) (*appv1.StatefulSet, error)
	Svc(context.Context) (*corev1.Service, error)
	SvcEndpoint() types.Endpoint
	Deploy(context.Context) (err error)

	GetConfig(ctx context.Context) (interface{}, error)
	Update(ctx context.Context, updateCfg interface{}) error
}

//// The following types are used for components without configuration files or implemation with other lanaguage

// ChainCoUpdate
type ChainCoConfig struct {
	Nodes     []string
	AuthUrl   string
	AuthToken string
}

type VenusWorkerConfig string //just mock here
