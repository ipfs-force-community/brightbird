package env

import (
	"context"
	"github.com/hunjixin/brightbird/types"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type IVenusDeployer IDeployer
type IChainCoDeployer IDeployer
type IMarketClientDeployer IDeployer
type IVenusAuthDeployer IDeployer
type IVenusGatewayDeployer IDeployer
type IVenusMarketDeployer IDeployer
type IVenusMessageDeployer IDeployer
type IVenusMinerDeployer IDeployer
type IVenusWalletDeployer IDeployer
type IVenusSectorManagerDeployer IDeployer
type IVenusWorkerDeployer IDeployer

type IDeployer interface {
	Name() string
	Pods() []corev1.Pod
	Deployment() []*appv1.Deployment
	Svc() *corev1.Service
	SvcEndpoint() types.Endpoint
	Deploy(ctx context.Context) (err error)
}
