package job

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/hunjixin/brightbird/env"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	yaml_k8s "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// TestRunnerDeployer used to deploy test runner
type TestRunnerDeployer struct {
	k8sClient *kubernetes.Clientset
	namespace string
	k8sCfg    *rest.Config
	dbs       []string
}

// NewK8sEnvDeployer create a new test environment
func NewTestRunnerDeployer(namespace string) (*TestRunnerDeployer, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		if errors.Is(err, rest.ErrNotInCluster) {
			var kubeConfig string
			if home := homedir.HomeDir(); home != "" {
				kubeConfig = filepath.Join(home, ".kube", "config")
			} else {
				return nil, errors.New("unable to get how path")
			}

			// use the current context in kubeConfig
			config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("unable to load config from incluster %w", err)
		}
	}

	// create the clientset
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &TestRunnerDeployer{
		k8sCfg:    config,
		k8sClient: k8sClient,
		namespace: namespace,
	}, nil
}

func (runnerDeployer *TestRunnerDeployer) ApplyRunner(ctx context.Context, f fs.File, args any) error {
	data, err := env.QuickRender(f, args)
	if err != nil {
		return err
	}

	deployment := &corev1.Pod{}
	err = yaml_k8s.Unmarshal(data, deployment)
	if err != nil {
		return err
	}
	log.Infof("runner config %s ...", string(data))
	name := deployment.Name
	podClient := runnerDeployer.k8sClient.CoreV1().Pods(runnerDeployer.namespace)
	log.Infof("Creating runner %s ...", name)
	_, err = podClient.Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	log.Infof("Created runner %s.", name)
	return nil
}

func (runnerDeployer *TestRunnerDeployer) GetLogs(ctx context.Context, testId string) error {
	//do nothing
	return nil
}

func (runnerDeployer *TestRunnerDeployer) CleanAll(ctx context.Context, testId string) error {
	//clean
	return nil
}
