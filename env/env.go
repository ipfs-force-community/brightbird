package env

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
	logging "github.com/ipfs/go-log/v2"
	"google.golang.org/appengine"
	"io"
	"io/fs"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	yaml_k8s "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/client-go/util/homedir"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var log = logging.Logger("env")

// CloseFunc use this to do some clean work after create a resource in k8s
type CloseFunc func() error

// JoinCloser wrap multiple closer to one
func JoinCloser(fns ...CloseFunc) CloseFunc {
	return func() error {
		mErr := appengine.MultiError{}
		for _, fn := range fns {
			if err := fn(); err != nil {
				mErr = append(mErr, err)
			}
		}

		if len(mErr) == 0 {
			return nil
		}
		return mErr
	}
}

// K8sEnvDeployer used to construct a k8s environment and do some k8s operation
type K8sEnvDeployer struct {
	k8sClient *kubernetes.Clientset
	namespace string
	hostIP    string
	testID    string
	k8sCfg    *rest.Config
	dialCtx   func(ctx context.Context, network, address string) (net.Conn, error)
	dbs       []string
}

// NewK8sEnvDeployer create a new test environment
func NewK8sEnvDeployer(namespace string, testID string) (*K8sEnvDeployer, error) {
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

	url, err := url.ParseRequestURI(config.Host)
	if err != nil {
		return nil, err
	}
	dialCtx := (&net.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}).DialContext
	return &K8sEnvDeployer{
		k8sCfg:    config,
		k8sClient: k8sClient,
		namespace: namespace,
		testID:    testID,
		hostIP:    url.Hostname(),
		dialCtx:   dialCtx,
	}, nil
}

// Debug return a unique test id
func (env *K8sEnvDeployer) TestID() string {
	return env.testID
}

// Debug return a unique test id
func (env *K8sEnvDeployer) UniqueId(outName string) string {
	if len(outName) > 0 {
		return env.testID + hex.EncodeToString(utils.Blake256([]byte(outName))[:4])
	}
	return env.testID
}

// RunDeployment deploy k8s's deployment from specific yaml config
func (env *K8sEnvDeployer) RunDeployment(ctx context.Context, f fs.File, args any) (*appv1.Deployment, error) {
	data, err := QuickRender(f, args)
	if err != nil {
		return nil, err
	}

	log.Debug("deployment yaml", string(data))
	deployment := &appv1.Deployment{}
	err = yaml_k8s.Unmarshal(data, deployment)
	if err != nil {
		return nil, err
	}

	if deployment.ObjectMeta.Labels == nil {
		deployment.ObjectMeta.Labels = map[string]string{"testid": env.TestID()}
		deployment.ObjectMeta.Labels = map[string]string{"apptype": "venus"}
	} else {
		deployment.ObjectMeta.Labels["testid"] = env.TestID()
		deployment.ObjectMeta.Labels["apptype"] = "venus"
	}

	name := deployment.Name
	deploymentClient := env.k8sClient.AppsV1().Deployments(env.namespace)
	log.Infof("Creating deployment %s ...", name)
	_, err = deploymentClient.Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	log.Infof("Created deployment %s.", name)

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancel when deploy %s", name)
		default:
			dep, err := deploymentClient.Get(ctx, deployment.GetName(), metav1.GetOptions{})
			if err != nil {
				return nil, err
			}

			replicas := int32(1)
			if deployment.Spec.Replicas != nil {
				replicas = *deployment.Spec.Replicas
			}
			if dep.Status.ReadyReplicas == replicas {
				return dep, nil
			}

			time.Sleep(time.Second * 5)
		}
	}
}

// RunDeployment deploy k8s's deployment from specific yaml config
func (env *K8sEnvDeployer) RunStatefulSets(ctx context.Context, f fs.File, args any) (*appv1.StatefulSet, error) {
	data, err := QuickRender(f, args)
	if err != nil {
		return nil, err
	}

	log.Debug("statefulset yaml", string(data))
	statefulSet := &appv1.StatefulSet{}
	err = yaml_k8s.Unmarshal(data, statefulSet)
	if err != nil {
		return nil, err
	}

	if statefulSet.ObjectMeta.Labels == nil {
		statefulSet.ObjectMeta.Labels = map[string]string{"testid": env.TestID()}
		statefulSet.ObjectMeta.Labels = map[string]string{"apptype": "venus"}
	} else {
		statefulSet.ObjectMeta.Labels["testid"] = env.TestID()
		statefulSet.ObjectMeta.Labels["apptype"] = "venus"
	}

	name := statefulSet.Name
	statefulSetClient := env.k8sClient.AppsV1().StatefulSets(env.namespace)
	log.Infof("Creating statefulSet %s ...", name)
	_, err = statefulSetClient.Create(ctx, statefulSet, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	log.Infof("Created statefulSet %s.\n", name)

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancel when deploy %s", name)
		default:
			dep, err := statefulSetClient.Get(ctx, statefulSet.GetName(), metav1.GetOptions{})
			if err != nil {
				return nil, err
			}

			if dep.Status.ReadyReplicas == *statefulSet.Spec.Replicas {
				return dep, nil
			}

			time.Sleep(time.Second * 5)
		}
	}
}

// RunService deploy k8s's service from specific yaml config
func (env *K8sEnvDeployer) RunService(ctx context.Context, fs fs.File, args any) (*corev1.Service, error) {
	data, err := QuickRender(fs, args)
	if err != nil {
		return nil, err
	}

	log.Debug("service yaml", string(data))
	serviceCfg := &corev1.Service{}
	err = yaml_k8s.Unmarshal(data, serviceCfg)
	if err != nil {
		return nil, err
	}

	if serviceCfg.ObjectMeta.Labels == nil {
		serviceCfg.ObjectMeta.Labels = map[string]string{"testid": env.TestID()}
		serviceCfg.ObjectMeta.Labels = map[string]string{"apptype": "venus"}
	} else {
		serviceCfg.ObjectMeta.Labels["testid"] = env.TestID()
		serviceCfg.ObjectMeta.Labels["apptype"] = "venus"
	}
	serviceClient := env.k8sClient.CoreV1().Services(env.namespace)
	log.Infof("Creating service %s ...", serviceCfg.GetName())
	result, err := serviceClient.Create(ctx, serviceCfg, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	log.Info("Created service %s", result.GetObjectMeta().GetName())
	return serviceClient.Get(ctx, serviceCfg.GetName(), metav1.GetOptions{})
}

func (env *K8sEnvDeployer) WaitForServiceReady(ctx context.Context, dep IDeployer) (types.Endpoint, error) {
	serviceClient := env.k8sClient.CoreV1().Services(env.namespace)
	name := dep.Svc().Name

	var endpoint types.Endpoint
LOOP:
	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("context cancel when deploy %s", name)
		default:
			service, err := serviceClient.Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return "", err
			}

			endpoints, err := env.k8sClient.CoreV1().Endpoints(env.namespace).Get(ctx, service.GetName(), metav1.GetOptions{})
			if err != nil {
				return "", err
			}
			if len(endpoints.Subsets) > 0 && len(endpoints.Subsets[0].Addresses) > 0 {
				if service.Spec.Type == corev1.ServiceTypeClusterIP {
					if service.Spec.ClusterIP == "None" {
						endpoint = types.Endpoint(fmt.Sprintf("%s:%d", name, service.Spec.Ports[0].Port))
						break LOOP
					} else {
						//todo check service is work
						if len(service.Spec.ClusterIP) > 0 {
							//take first
							endpoint = types.Endpoint(fmt.Sprintf("%s:%d", service.Spec.ClusterIP, service.Spec.Ports[0].Port))
							break LOOP
						} else {
							return "", fmt.Errorf("unable to get cluser ip for %s", name)
						}
					}
				}
				return "", fmt.Errorf("unable service type %s(%s)", name, service.Spec.Type)
			}
			continue
		}
	}

	var err error
	if Debug {
		//forward a pod to local machine
		for {
			//pod network may not be ready
			forwardPort, err := env.PortForwardPod(ctx, dep.Pods()[0].GetName(), int(dep.Svc().Spec.Ports[0].Port))
			if err != nil {

				return "", err
			}
			err = env.WaitForAPIReady(ctx, forwardPort)
			if err != nil {
				if errors.Is(err, io.EOF) {
					time.Sleep(time.Second * 5)
					continue
				}
				return "", err
			}
			break
		}
	} else {
		for {
			err = env.WaitEndpointRead(ctx, endpoint)
			if err != nil {
				return "", err
			}
			err = env.WaitForAPIReady(ctx, endpoint)
			if err != nil {
				if errors.Is(err, io.EOF) {
					time.Sleep(time.Second * 5)
					continue
				}
				return "", err
			}
		}
	}

	return endpoint, nil
}

func (env *K8sEnvDeployer) CreateConfigMap(ctx context.Context, f fs.File, args any) (*corev1.ConfigMap, error) {
	data, err := QuickRender(f, args)
	if err != nil {
		return nil, err
	}

	log.Debug("configmap yaml", string(data))
	configMap := &corev1.ConfigMap{}
	err = yaml_k8s.Unmarshal(data, configMap)
	if err != nil {
		return nil, err
	}

	if configMap.ObjectMeta.Labels == nil {
		configMap.ObjectMeta.Labels = map[string]string{"testid": env.TestID()}
		configMap.ObjectMeta.Labels = map[string]string{"apptype": "venus"}
	} else {
		configMap.ObjectMeta.Labels["testid"] = env.TestID()
		configMap.ObjectMeta.Labels["apptype"] = "venus"
	}

	configMapClient := env.k8sClient.CoreV1().ConfigMaps(env.namespace)
	log.Infof("Creating configmap %s ...", configMap.GetName())
	result, err := configMapClient.Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	log.Infof("Created configmap %s.", result.GetObjectMeta().GetName())
	return configMap, nil
}

func (env *K8sEnvDeployer) WaitForAPIReady(ctx context.Context, endpoint types.Endpoint) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancel")
		default:
			tCtx, _ := context.WithTimeout(ctx, time.Second*5)
			req, err := http.NewRequestWithContext(tCtx, "GET", "http://"+string(endpoint)+"/healthcheck", nil)
			if err != nil {
				return err
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			if resp.StatusCode == http.StatusOK {
				return nil
			}
			log.Debugf("track status %s %d", resp.Status, resp.StatusCode)
			time.Sleep(time.Second * 5)
			continue
		}
	}
}

func (env *K8sEnvDeployer) GetSvcEndpoint(svc *corev1.Service) (string, error) {
	if svc.Spec.Type == corev1.ServiceTypeClusterIP {
		if svc.Spec.ClusterIP == "None" {
			return fmt.Sprintf("%s:%d", svc.GetName(), svc.Spec.Ports[0].Port), nil
		} else {
			//todo check service is work
			if len(svc.Spec.ClusterIP) > 0 {
				//take first
				return fmt.Sprintf("%s:%d", svc.Spec.ClusterIP, svc.Spec.Ports[0].Port), nil
			} else {
				return "", fmt.Errorf("unable to get cluser ip for %s", svc.GetName())
			}
		}
	} else if svc.Spec.Type == corev1.ServiceTypeNodePort {
		return fmt.Sprintf("%s:%d", env.hostIP, svc.Spec.Ports[0].Port), nil
	}
	return "", fmt.Errorf("not support service type %s", svc.GetName())
}

func (env *K8sEnvDeployer) GetPodsByLabel(ctx context.Context, deployAppLabel string) ([]corev1.Pod, error) {
	podClient := env.k8sClient.CoreV1().Pods(env.namespace)
	podList, err := podClient.List(ctx, metav1.ListOptions{LabelSelector: "app=" + deployAppLabel})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

// ReadSmallFilelInPod read small file content from pod, dont not use this function to read big file
func (env *K8sEnvDeployer) ReadSmallFilelInPod(ctx context.Context, podName string, path string) ([]byte, error) {
	cmd := []string{
		"cat",
		path,
	}
	req := env.k8sClient.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace(env.namespace).SubResource("exec")
	option := &corev1.PodExecOptions{
		Command: cmd,
		Stdin:   false,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(env.k8sCfg, "POST", req.URL())
	if err != nil {
		return nil, err
	}
	stdOut := bytes.NewBuffer(nil)
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: stdOut,
		Stderr: os.Stderr,
		Tty:    true,
	})
	if err != nil {
		return nil, err
	}
	return io.ReadAll(stdOut)
}

func (env *K8sEnvDeployer) CreateDatabase(dsn string) error {
	err := utils.CreateDatabase(dsn)
	if err != nil {
		env.dbs = append(env.dbs, dsn)
	}
	return err
}

func (env *K8sEnvDeployer) WaitEndpointRead(ctx context.Context, endpoint types.Endpoint) error {
	for {
		select {
		case <-ctx.Done():
			return errors.New("context cancel")
		default:
			tCtx, _ := context.WithTimeout(ctx, time.Second*5)
			_, err := env.dialCtx(tCtx, "tcp", string(endpoint))
			if err == nil {
				return err
			}
		}
	}
}

// PortForwardPod forward pod api service to local machine, used for debug
func (env *K8sEnvDeployer) PortForwardPod(ctx context.Context, podName string, destPort int) (types.Endpoint, error) {
	readyCh := make(chan struct{})
	stopCh := make(chan struct{})
	reqURL := env.k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(env.namespace).
		Name(podName).
		SubResource("portforward").URL()
	transport, upgrader, err := spdy.RoundTripperFor(env.k8sCfg)
	if err != nil {
		return "", err
	}

	freePort, err := utils.GetFreePort()
	if err != nil {
		return "", err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, reqURL)
	fw, err := portforward.NewOnAddresses(dialer, []string{"127.0.0.1"}, []string{strconv.Itoa(freePort) + ":" + strconv.Itoa(destPort)}, stopCh, readyCh, os.Stdout, os.Stdout)
	if err != nil {
		return "", err
	}

	go func() {
		err = fw.ForwardPorts()
		if err != nil {
			log.Errorf("forward port error %v", err)
		}
	}()

	go func() {
		select {
		case <-ctx.Done():
			stopCh <- struct{}{}
		}
	}()

	select {
	case <-ctx.Done():
		return "", errors.New("context cancel")
	case <-readyCh:
	}

	return types.EndpointFromHostPort("127.0.0.1", freePort), nil
}

func (env *K8sEnvDeployer) Clean(ctx context.Context) error {
	err := env.k8sClient.AppsV1().Deployments(env.namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: "testid=" + env.TestID()})
	if err != nil {
		log.Errorf("clean deployment failed %s", err)
	}
	err = env.k8sClient.AppsV1().StatefulSets(env.namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: "testid=" + env.TestID()})
	if err != nil {
		log.Errorf("clean statefuleset failed %s", err)
	}
	services, err := env.k8sClient.CoreV1().Services(env.namespace).List(ctx, metav1.ListOptions{LabelSelector: "testid=" + env.TestID()})
	if err != nil {
		log.Errorf("get service failed %s", err)
	}
	for _, svc := range services.Items {
		err := env.k8sClient.CoreV1().Services(env.namespace).Delete(ctx, svc.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Errorf("delete service failed %s", err)
		}
	}

	for _, dsn := range env.dbs {
		err = utils.DropDatabase(dsn)
		if err != nil {
			log.Errorf("drop %s failed %s", dsn, err)
		}
	}
	return nil
}
