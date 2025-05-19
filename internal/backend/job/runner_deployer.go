package job

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptrace"
	"net/textproto"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	yaml_k8s "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/ipfs-force-community/brightbird/pkg/env"
)

// LoggingRoundTripper 实现了 http.RoundTripper 接口，用于记录请求和响应
type LoggingRoundTripper struct {
	rt http.RoundTripper
}

// RoundTrip 记录请求和响应
func (lrt *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Printf("Request: %s %s\n", req.Method, req.URL.String())
	for key, values := range req.Header.Clone() {
		for _, value := range values {
			fmt.Printf("Header: %s: %s\n", key, value)
		}
	}
	var start time.Time
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), &httptrace.ClientTrace{
		DNSStart: func(ds httptrace.DNSStartInfo) {
			fmt.Println("DNS Lookup Start...")
			start = time.Now()
		},
		DNSDone: func(di httptrace.DNSDoneInfo) {
			fmt.Printf("DNS Lookup Done. Took: %s\n", time.Since(start))
		},
		ConnectStart: func(network, addr string) {
			fmt.Println("Dialing...")
			start = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			fmt.Printf("Dialed. Took: %s\n", time.Since(start))
		},
		TLSHandshakeStart: func() {
			fmt.Println("TLS Handshake Start...")
			start = time.Now()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			fmt.Printf("TLS Handshake Done. Took: %s\n", time.Since(start))
			if err != nil {
				fmt.Println("TLS Handshake err: ", err)
			}
		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			fmt.Printf("Got Conn: %+v\n", connInfo)
			start = time.Now()
		},
		GotFirstResponseByte: func() {
			fmt.Printf("Got First Response Byte. Took: %s\n", time.Since(start))
		},
		GetConn: func(hostPort string) {
			fmt.Println("GetConn ", hostPort)
		},
		PutIdleConn: func(err error) {
			fmt.Println("PutIdleConn ", err)
		},
		Got100Continue: func() {
			fmt.Println("Got 100 Continue")
		},
		Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
			fmt.Printf("Got 1xx Response: %d\n", code)
			return nil
		},
		WroteHeaderField: func(key string, value []string) {
			fmt.Printf("Wrote Header Field: %s: %s\n", key, value)
		},
		WroteHeaders: func() {
			fmt.Println("Wrote Headers")
		},
		Wait100Continue: func() {
			fmt.Println("Wait 100 Continue")
		},
		WroteRequest: func(rInfo httptrace.WroteRequestInfo) {
			fmt.Println("Wrote Request")
			if rInfo.Err != nil {
				fmt.Println("WroteRequest err: ", rInfo.Err)
			}
		},
	}))

	resp, err := lrt.rt.RoundTrip(req)
	if err == nil {
		fmt.Printf("[ok]Response: %s %s - %d\n", req.Method, req.URL.String(), resp.StatusCode)
	} else {
		fmt.Printf("[error]Response: %s %s - %v\n", req.Method, req.URL.String(), err)
	}
	return resp, err
}

// NewLoggingRoundTripper 创建一个新的 LoggingRoundTripper
func NewLoggingRoundTripper(rt http.RoundTripper) *LoggingRoundTripper {
	return &LoggingRoundTripper{rt: rt}
}

// TestRunnerDeployer used to deploy test scriptRunner
type TestRunnerDeployer struct {
	k8sClient *kubernetes.Clientset
	namespace string
	mysql     string
	k8sCfg    *rest.Config
}

// NewK8sEnvDeployer create a new test environment
func NewTestRunnerDeployer(namespace string, mysql string) (*TestRunnerDeployer, error) {
	kubeConfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	if _, err := os.Stat(kubeConfig); err != nil {
		kubeConfig = ""
	} else {
		fmt.Println("load kubeConfig from : ", kubeConfig)
	}
	// use the current context in kubeConfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to load config: %w", err)
	}

	// // for trace http request
	// config.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
	// 	return NewLoggingRoundTripper(rt)
	// }

	// create the clientset
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &TestRunnerDeployer{
		k8sCfg:    config,
		k8sClient: k8sClient,
		namespace: namespace,
		mysql:     mysql,
	}, nil
}

func (runnerDeployer *TestRunnerDeployer) ApplyRunner(ctx context.Context, f fs.File, args any) (*corev1.Pod, error) {
	data, err := env.QuickRender(f, args)
	if err != nil {
		return nil, err
	}
	jobLog.Infof("scriptRunner config :\n%s", string(data))
	deployment := &corev1.Pod{}
	err = yaml_k8s.Unmarshal(data, deployment)
	if err != nil {
		return nil, err
	}
	name := deployment.Name
	podClient := runnerDeployer.k8sClient.CoreV1().Pods(runnerDeployer.namespace)
	jobLog.Infof("Creating scriptRunner %s ...", name)
	if _, err = podClient.Create(ctx, deployment, metav1.CreateOptions{}); err != nil {
		return nil, err
	}
	jobLog.Infof("Created scriptRunner %s.", name)

	pod, err := podClient.Get(ctx, deployment.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return pod, nil
}

func (runnerDeployer *TestRunnerDeployer) CheckTestRunner(ctx context.Context, id string) (int, error) {
	podClient := runnerDeployer.k8sClient.CoreV1().Pods(runnerDeployer.namespace)
	pod, err := podClient.Get(ctx, id, metav1.GetOptions{})
	if err != nil {
		return 0, err
	}
	if pod.Status.Phase == corev1.PodFailed {
		return 0, fmt.Errorf("runnerpod error %v", pod.Status.Message)
	}
	for _, container := range pod.Status.ContainerStatuses {
		if container.LastTerminationState.Terminated != nil && container.LastTerminationState.Terminated.ExitCode != 0 {
			return int(container.RestartCount), fmt.Errorf("runner pod crashed %v", pod.Status.Message)
		}
	}
	return 0, nil
}

func (runnerDeployer *TestRunnerDeployer) GetLogs(ctx context.Context, testId string) error {
	//do nothing
	return nil
}

func (runnerDeployer *TestRunnerDeployer) RemoveFinishRunner(ctx context.Context) error {
	//clean scriptRunner status.phase==Succeeded
	return runnerDeployer.k8sClient.CoreV1().Pods(runnerDeployer.namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{FieldSelector: "status.phase==Succeeded"})
}

func (runnerDeployer *TestRunnerDeployer) CleanTestResource(ctx context.Context, testId string) error {
	resourceMg := env.NewResourceMgr(runnerDeployer.k8sClient, runnerDeployer.namespace, runnerDeployer.mysql, testId)
	return resourceMg.Clean(ctx)
}
