package venus_wallet

import (
	"context"
	"reflect"
	"testing"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/types"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestDefaultConfig(t *testing.T) {
	tests := []struct {
		name string
		want Config
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DefaultConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewVenusWalletDeployer(t *testing.T) {
	type args struct {
		env             *env.K8sEnvDeployer
		gatewayUrl      string
		userToken       string
		supportAccounts []string
	}
	tests := []struct {
		name string
		args args
		want *VenusWalletDeployer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewVenusWalletDeployer(tt.args.env, tt.args.gatewayUrl, tt.args.userToken, tt.args.supportAccounts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewVenusWalletDeployer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeployerFromConfig(t *testing.T) {
	type args struct {
		env    *env.K8sEnvDeployer
		cfg    Config
		params Config
	}
	tests := []struct {
		name    string
		args    args
		want    env.IDeployer
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeployerFromConfig(tt.args.env, tt.args.cfg, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeployerFromConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeployerFromConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVenusWalletDeployer_Name(t *testing.T) {
	type fields struct {
		env             *env.K8sEnvDeployer
		cfg             *Config
		svcEndpoint     types.Endpoint
		configMapName   string
		statefulSetName string
		svcName         string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployer := &VenusWalletDeployer{
				env:             tt.fields.env,
				cfg:             tt.fields.cfg,
				svcEndpoint:     tt.fields.svcEndpoint,
				configMapName:   tt.fields.configMapName,
				statefulSetName: tt.fields.statefulSetName,
				svcName:         tt.fields.svcName,
			}
			if got := deployer.Name(); got != tt.want {
				t.Errorf("VenusWalletDeployer.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVenusWalletDeployer_Pods(t *testing.T) {
	type fields struct {
		env             *env.K8sEnvDeployer
		cfg             *Config
		svcEndpoint     types.Endpoint
		configMapName   string
		statefulSetName string
		svcName         string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []corev1.Pod
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployer := &VenusWalletDeployer{
				env:             tt.fields.env,
				cfg:             tt.fields.cfg,
				svcEndpoint:     tt.fields.svcEndpoint,
				configMapName:   tt.fields.configMapName,
				statefulSetName: tt.fields.statefulSetName,
				svcName:         tt.fields.svcName,
			}
			got, err := deployer.Pods(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("VenusWalletDeployer.Pods() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VenusWalletDeployer.Pods() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVenusWalletDeployer_StatefulSet(t *testing.T) {
	type fields struct {
		env             *env.K8sEnvDeployer
		cfg             *Config
		svcEndpoint     types.Endpoint
		configMapName   string
		statefulSetName string
		svcName         string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *appv1.StatefulSet
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployer := &VenusWalletDeployer{
				env:             tt.fields.env,
				cfg:             tt.fields.cfg,
				svcEndpoint:     tt.fields.svcEndpoint,
				configMapName:   tt.fields.configMapName,
				statefulSetName: tt.fields.statefulSetName,
				svcName:         tt.fields.svcName,
			}
			got, err := deployer.StatefulSet(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("VenusWalletDeployer.StatefulSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VenusWalletDeployer.StatefulSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVenusWalletDeployer_Svc(t *testing.T) {
	type fields struct {
		env             *env.K8sEnvDeployer
		cfg             *Config
		svcEndpoint     types.Endpoint
		configMapName   string
		statefulSetName string
		svcName         string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *corev1.Service
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployer := &VenusWalletDeployer{
				env:             tt.fields.env,
				cfg:             tt.fields.cfg,
				svcEndpoint:     tt.fields.svcEndpoint,
				configMapName:   tt.fields.configMapName,
				statefulSetName: tt.fields.statefulSetName,
				svcName:         tt.fields.svcName,
			}
			got, err := deployer.Svc(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("VenusWalletDeployer.Svc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VenusWalletDeployer.Svc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVenusWalletDeployer_SvcEndpoint(t *testing.T) {
	type fields struct {
		env             *env.K8sEnvDeployer
		cfg             *Config
		svcEndpoint     types.Endpoint
		configMapName   string
		statefulSetName string
		svcName         string
	}
	tests := []struct {
		name   string
		fields fields
		want   types.Endpoint
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployer := &VenusWalletDeployer{
				env:             tt.fields.env,
				cfg:             tt.fields.cfg,
				svcEndpoint:     tt.fields.svcEndpoint,
				configMapName:   tt.fields.configMapName,
				statefulSetName: tt.fields.statefulSetName,
				svcName:         tt.fields.svcName,
			}
			if got := deployer.SvcEndpoint(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VenusWalletDeployer.SvcEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVenusWalletDeployer_Deploy(t *testing.T) {
	type fields struct {
		env             *env.K8sEnvDeployer
		cfg             *Config
		svcEndpoint     types.Endpoint
		configMapName   string
		statefulSetName string
		svcName         string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployer := &VenusWalletDeployer{
				env:             tt.fields.env,
				cfg:             tt.fields.cfg,
				svcEndpoint:     tt.fields.svcEndpoint,
				configMapName:   tt.fields.configMapName,
				statefulSetName: tt.fields.statefulSetName,
				svcName:         tt.fields.svcName,
			}
			if err := deployer.Deploy(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("VenusWalletDeployer.Deploy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVenusWalletDeployer_GetConfig(t *testing.T) {
	type fields struct {
		env             *env.K8sEnvDeployer
		cfg             *Config
		svcEndpoint     types.Endpoint
		configMapName   string
		statefulSetName string
		svcName         string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployer := &VenusWalletDeployer{
				env:             tt.fields.env,
				cfg:             tt.fields.cfg,
				svcEndpoint:     tt.fields.svcEndpoint,
				configMapName:   tt.fields.configMapName,
				statefulSetName: tt.fields.statefulSetName,
				svcName:         tt.fields.svcName,
			}
			got, err := deployer.GetConfig(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("VenusWalletDeployer.GetConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VenusWalletDeployer.GetConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVenusWalletDeployer_Update(t *testing.T) {
	type fields struct {
		env             *env.K8sEnvDeployer
		cfg             *Config
		svcEndpoint     types.Endpoint
		configMapName   string
		statefulSetName string
		svcName         string
	}
	type args struct {
		ctx       context.Context
		updateCfg interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployer := &VenusWalletDeployer{
				env:             tt.fields.env,
				cfg:             tt.fields.cfg,
				svcEndpoint:     tt.fields.svcEndpoint,
				configMapName:   tt.fields.configMapName,
				statefulSetName: tt.fields.statefulSetName,
				svcName:         tt.fields.svcName,
			}
			if err := deployer.Update(tt.args.ctx, tt.args.updateCfg); (err != nil) != tt.wantErr {
				t.Errorf("VenusWalletDeployer.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
