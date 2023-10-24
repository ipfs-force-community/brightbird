package env

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/utils"
	logging "github.com/ipfs/go-log/v2"
	"github.com/pelletier/go-toml/v2"
	"google.golang.org/appengine"
	"gopkg.in/yaml.v3"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	yaml_k8s "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/client-go/util/homedir"
)

var log = logging.Logger("env")

// CloseFunc use this to do some clean work after create a resource in k8s
type CloseFunc func() error

type PodGetter func(ctx context.Context, k8sEnv *K8sEnvDeployer) ([]corev1.Pod, error)

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
	k8sClient         *kubernetes.Clientset
	namespace         string
	hostIP            string
	testID            string
	retry             int
	registry          string
	mysqlConnTemplate string
	k8sCfg            *rest.Config
	dialCtx           func(ctx context.Context, network, address string) (net.Conn, error)
	resourceMgr       IResourceMgr
}

type K8sInitParams struct {
	Namespace         string `json:"namespace"`
	TestID            string `json:"testID"`
	Retry             int    `json:"retry"`
	Registry          string `json:"registry"`
	MysqlConnTemplate string `json:"mysqlConnTemplate"`
}

func K8sFromConfig() (*rest.Config, error) {
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
	return config, nil
}

// NewK8sEnvDeployer create a new test environment
func NewK8sEnvDeployer(params K8sInitParams) (*K8sEnvDeployer, error) {
	config, err := K8sFromConfig()
	if err != nil {
		return nil, err
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
		k8sCfg:            config,
		k8sClient:         k8sClient,
		namespace:         params.Namespace,
		testID:            params.TestID,
		retry:             params.Retry,
		hostIP:            url.Hostname(),
		dialCtx:           dialCtx,
		registry:          params.Registry,
		mysqlConnTemplate: params.MysqlConnTemplate,
		resourceMgr:       NewResourceMgr(k8sClient, params.Namespace, params.MysqlConnTemplate, params.TestID),
	}, nil
}

// TestID return a unique test id
func (env *K8sEnvDeployer) FormatMysqlConnection(dbName string) string {
	return fmt.Sprintf(env.mysqlConnTemplate, dbName)
}

// TestID return a resource id
func (env *K8sEnvDeployer) ResourceMgr() IResourceMgr {
	return env.resourceMgr
}

// TestID return a unique test id
func (env *K8sEnvDeployer) TestID() string {
	return env.testID
}

// Retry return task retry
func (env *K8sEnvDeployer) Retry() int {
	return env.retry
}

// Registry
func (env *K8sEnvDeployer) Registry() string {
	return env.registry
}

// NameSpace
func (env *K8sEnvDeployer) NameSpace() string {
	return env.namespace
}

// K8sClient
func (env *K8sEnvDeployer) K8sClient() *kubernetes.Clientset {
	return env.k8sClient
}

func (env *K8sEnvDeployer) setCommonLabels(objectMeta *metav1.ObjectMeta) {
	if objectMeta.Labels == nil {
		objectMeta.Labels = map[string]string{}
	}
	objectMeta.Namespace = env.namespace
	objectMeta.Labels["testid"] = env.TestID()
	objectMeta.Labels["retry"] = strconv.Itoa(env.Retry())
	objectMeta.Labels["apptype"] = "venus"
}

func (env *K8sEnvDeployer) setPrivateRegistry(statefulSet *corev1.PodTemplateSpec) {
	for _, c := range statefulSet.Spec.Containers {
		if len(env.registry) > 0 {
			c.Image = fmt.Sprintf("%s/%s", env.registry, c.Image)
		}
	}

	for _, c := range statefulSet.Spec.InitContainers {
		if len(env.registry) > 0 {
			c.Image = fmt.Sprintf("%s/%s", env.registry, c.Image)
		}
	}
}

func (env *K8sEnvDeployer) StopPods(ctx context.Context, pods []corev1.Pod) error {
	for _, pod := range pods {
		err := env.k8sClient.CoreV1().Pods(env.namespace).Delete(ctx, pod.GetName(), metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (env *K8sEnvDeployer) UpdateConfigMaps(ctx context.Context, configMap *corev1.ConfigMap) error {
	configMapClient := env.k8sClient.CoreV1().ConfigMaps(env.namespace)
	log.Infof("Try to update %s ", configMap.GetName())
	_, err := configMapClient.Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("update configMap(%s) fail %w", configMap.GetName(), err)
	}
	log.Infof("Updated configMap %s.", configMap.GetName())
	return nil
}

func (env *K8sEnvDeployer) UpdateStatefulSets(ctx context.Context, statefulset *appv1.StatefulSet) error {
	statefulSetClient := env.k8sClient.AppsV1().StatefulSets(env.namespace)
	log.Infof("Try to update %s ", statefulset.GetName())

	// 更新 StatefulSet
	updatedStatefulSet, err := statefulSetClient.Update(ctx, statefulset, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("update statefulset(%s) %w", statefulset.GetName(), err)
	}
	// 检查更新是否成功
	err = env.waitForStatefulSetUpdate(ctx, updatedStatefulSet)
	if err != nil {
		return err
	}
	log.Infof("Updated statefulSet %s.", statefulset.GetName())

	return nil
}

func (env *K8sEnvDeployer) waitForStatefulSetUpdate(ctx context.Context, statefulset *appv1.StatefulSet) error {
	statefulSetClient := env.k8sClient.AppsV1().StatefulSets(env.namespace)

	// 设置超时和轮询间隔
	timeout := time.Minute * 5
	pollInterval := time.Second * 5
	deadline := time.Now().Add(timeout)

	// 循环检查 StatefulSet 是否更新成功
	for {
		// 获取最新的 StatefulSet
		currentStatefulSet, err := statefulSetClient.Get(ctx, statefulset.GetName(), metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get updated StatefulSet: %w", err)
		}

		// 检查 observedGeneration 是否与更新前的一致
		if currentStatefulSet.Status.ObservedGeneration == statefulset.Status.ObservedGeneration {
			// 如果 observedGeneration 未增加，继续等待
			if time.Now().After(deadline) {
				return fmt.Errorf("timed out waiting for StatefulSet to update")
			}
			time.Sleep(pollInterval)
		} else {
			// 更新已成功
			break
		}
	}

	return nil
}

func (env *K8sEnvDeployer) UpdateStatefulSetsByName(ctx context.Context, stateName string) error {
	statefulSetClient := env.k8sClient.AppsV1().StatefulSets(env.namespace)
	statefulSet, err := statefulSetClient.Get(ctx, stateName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get statefulset(%s) fail %w", stateName, err)
	}

	return env.UpdateStatefulSets(ctx, statefulSet)
}

func (env *K8sEnvDeployer) DeletePodAndWait(ctx context.Context, podName string) error {
	podClient := env.k8sClient.CoreV1().Pods(env.namespace)
	err := podClient.Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return wait.PollImmediateUntilWithContext(ctx, time.Second*3, func(ctx context.Context) (done bool, err error) {
		pod, err := podClient.Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			if errors2.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}

		return pod.DeletionTimestamp == nil && pod.Status.Phase == corev1.PodRunning && pod.Status.ContainerStatuses[0].Ready, nil
	})
}

func (env *K8sEnvDeployer) WaitPodReady(ctx context.Context, podName string) error {
	podClient := env.k8sClient.CoreV1().Pods(env.namespace)
	return wait.PollImmediateUntilWithContext(ctx, time.Second*3, func(ctx context.Context) (done bool, err error) {
		pod, err := podClient.Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			if errors2.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}

		return pod.Status.Phase == corev1.PodRunning && pod.Status.ContainerStatuses[0].Ready, nil
	})
}

// CreatePvc deploy k8s's deployment from specific yaml config
func (env *K8sEnvDeployer) CreatePvc(ctx context.Context, f fs.File, args any) (*corev1.PersistentVolumeClaim, error) {
	data, err := QuickRender(f, args)
	if err != nil {
		return nil, fmt.Errorf("render pvc fail %w", err)
	}

	pvc := &corev1.PersistentVolumeClaim{}
	err = yaml_k8s.Unmarshal(data, pvc)
	if err != nil {
		fmt.Println(string(data))
		return nil, fmt.Errorf("unmarshal to pvc fail %w", err)
	}
	env.setCommonLabels(&pvc.ObjectMeta)

	cfgData, err := yaml.Marshal(pvc)
	if err != nil {
		return nil, err
	}

	name := pvc.Name
	log.Debugf("pvc(%s) yaml config %s", pvc.GetName(), string(cfgData))
	pvc, err = env.k8sClient.CoreV1().PersistentVolumeClaims(env.namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("create pvc %s fail", name)
	}

	log.Infof("Created pvc %s.", name)
	return pvc, nil
}

// RunStatefulSets deploy k8s's deployment from specific yaml config
func (env *K8sEnvDeployer) RunStatefulSets(ctx context.Context, podGetter PodGetter, f fs.File, args any) (*appv1.StatefulSet, error) {
	data, err := QuickRender(f, args)
	if err != nil {
		return nil, fmt.Errorf("render statefulset fail %w", err)
	}

	statefulSet := &appv1.StatefulSet{}
	err = yaml_k8s.Unmarshal(data, statefulSet)
	if err != nil {
		fmt.Println(string(data))
		return nil, fmt.Errorf("unmarshal to statefulset fail %w", err)
	}

	env.setCommonLabels(&statefulSet.ObjectMeta)
	env.setCommonLabels(&statefulSet.Spec.Template.ObjectMeta)
	env.setPrivateRegistry(&statefulSet.Spec.Template)
	for _, pvc := range statefulSet.Spec.VolumeClaimTemplates {
		env.setCommonLabels(&pvc.ObjectMeta)
	}

	cfgData, err := yaml.Marshal(statefulSet)
	if err != nil {
		return nil, err
	}
	log.Debugf("statefulset(%s) yaml config %s", statefulSet.GetName(), string(cfgData))

	name := statefulSet.Name
	statefulSetClient := env.k8sClient.AppsV1().StatefulSets(env.namespace)

	_, err = statefulSetClient.Get(ctx, statefulSet.GetName(), metav1.GetOptions{})
	if err != nil {
		if errors2.IsNotFound(err) {
			log.Infof("Creating statefulSet %s ...", name)
			_, err = statefulSetClient.Create(ctx, statefulSet, metav1.CreateOptions{})
			if err != nil {
				return nil, fmt.Errorf("create statefulset(%s) fail %w", name, err)
			}
			log.Infof("Created statefulSet %s.", name)
		} else {
			return nil, err
		}
	} else {
		log.Infof("Statefulset already exit try to update %s ", statefulSet.GetName())
		_, err = statefulSetClient.Update(ctx, statefulSet, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("update statefulset(%s) fail %w", name, err)
		}
		log.Infof("Updated statefulSet %s.", name)
	}

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancel when deploy %s", name)
		case <-ticker.C:
			dep, err := statefulSetClient.Get(ctx, statefulSet.GetName(), metav1.GetOptions{})
			if err != nil {
				if errors2.IsNotFound(err) {
					continue
				}
				return nil, fmt.Errorf("get statefulset(%s) fail %w", statefulSet.GetName(), err)
			}

			if dep.Status.ReadyReplicas == *statefulSet.Spec.Replicas {
				return dep, nil
			}

			//detect crash
			pods, err := podGetter(ctx, env)
			if err != nil {
				return nil, err
			}

			time.Sleep(time.Second * 5)
			for _, pod := range pods {
				if len(pod.Status.ContainerStatuses) > 0 && pod.Status.ContainerStatuses[0].RestartCount > 1 {
					//pod maybe crash
					return nil, fmt.Errorf("pod restart more than 1 times, maybe crash")
				}
			}
		}
	}
}

// RunConfigMap create config map for app
func (env *K8sEnvDeployer) RunConfigMap(ctx context.Context, f fs.File, args any) (*corev1.ConfigMap, error) {
	data, err := QuickRender(f, args)
	if err != nil {
		return nil, fmt.Errorf("render configmap fail %w", err)
	}

	configMap := &corev1.ConfigMap{}
	err = yaml_k8s.Unmarshal(data, configMap)
	if err != nil {
		return nil, fmt.Errorf("unmarshal to configmap fail %w", err)
	}

	env.setCommonLabels(&configMap.ObjectMeta)
	log.Infof("configmap(%s) yaml config %s", configMap.GetName(), string(data))

	configMapClient := env.k8sClient.CoreV1().ConfigMaps(env.namespace)
	name := configMap.GetName()
	_, err = configMapClient.Get(ctx, configMap.GetName(), metav1.GetOptions{})
	if err != nil {
		if errors2.IsNotFound(err) {
			log.Infof("Creating configmap %s ...", name)
			_, err := configMapClient.Create(ctx, configMap, metav1.CreateOptions{})
			if err != nil {
				return nil, fmt.Errorf("create configmap(%s) fail %w", name, err)
			}
			log.Infof("Created configmap %s.", name)
		} else {
			return nil, err
		}
	} else {
		log.Infof("ConfigMap already exit try to update %s ", name)
		_, err = configMapClient.Update(ctx, configMap, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("update configmap(%s) fail %w", name, err)
		}
	}

	return configMapClient.Get(ctx, configMap.GetName(), metav1.GetOptions{})
}

// RunService deploy k8s's service from specific yaml config
func (env *K8sEnvDeployer) RunService(ctx context.Context, fs fs.File, args any) (*corev1.Service, error) {
	data, err := QuickRender(fs, args)
	if err != nil {
		return nil, fmt.Errorf("render service fail %w", err)
	}

	serviceCfg := &corev1.Service{}
	err = yaml_k8s.Unmarshal(data, serviceCfg)
	if err != nil {
		return nil, fmt.Errorf("unmarshal to service fail %w", err)
	}

	env.setCommonLabels(&serviceCfg.ObjectMeta)
	cfgData, err := yaml.Marshal(serviceCfg)
	if err != nil {
		return nil, err
	}

	svcName := serviceCfg.GetName()
	log.Debugf("service(%s) yaml config %s", svcName, string(cfgData))

	serviceClient := env.k8sClient.CoreV1().Services(env.namespace)
	log.Infof("Creating service %s ...", svcName)
	_, err = serviceClient.Create(ctx, serviceCfg, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("create service(%s) fail %w", svcName, err)
	}
	log.Infof("Created service %s", svcName)
	return serviceClient.Get(ctx, svcName, metav1.GetOptions{})
}

func (env *K8sEnvDeployer) WaitForServiceReady(ctx context.Context, svc *corev1.Service, healthchecker func(context.Context, types.Endpoint) error) (types.Endpoint, error) {
	serviceClient := env.k8sClient.CoreV1().Services(env.namespace)
	name := svc.GetName()

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	var endpoint types.Endpoint
LOOP:
	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("context cancel when deploy %s", name)
		case <-ticker.C:
			service, err := serviceClient.Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				if errors2.IsNotFound(err) {
					continue
				}
				return "", fmt.Errorf("get service of %s fail %w", name, err)
			}

			endpoints, err := env.k8sClient.CoreV1().Endpoints(env.namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				if errors2.IsNotFound(err) {
					continue
				}
				return "", fmt.Errorf("get endpoint of %s fail %w", name, err)
			}
			log.Infof("service %v", service.Spec)
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

	err := env.WaitEndpointReady(ctx, endpoint)
	if err != nil {
		return "", err
	}

	log.Infof("use cluster ip %s", endpoint)

	if healthchecker != nil {
		err = healthchecker(ctx, endpoint)
		if err != nil {
			return "", err
		}
	}

	return endpoint, nil
}

func (env *K8sEnvDeployer) GetSvcEndpoint(svc *corev1.Service) (string, error) {
	if svc.Spec.Type == corev1.ServiceTypeClusterIP {
		if svc.Spec.ClusterIP == "None" {
			return fmt.Sprintf("%s:%d", svc.GetName(), svc.Spec.Ports[0].Port), nil
		}

		//todo check service is work
		if len(svc.Spec.ClusterIP) > 0 {
			//take first
			return fmt.Sprintf("%s:%d", svc.Spec.ClusterIP, svc.Spec.Ports[0].Port), nil
		}
		return "", fmt.Errorf("unable to get cluser ip for %s", svc.GetName())
	} else if svc.Spec.Type == corev1.ServiceTypeNodePort {
		return fmt.Sprintf("%s:%d", env.hostIP, svc.Spec.Ports[0].Port), nil
	}
	return "", fmt.Errorf("not support service type %s", svc.GetName())
}

func (env *K8sEnvDeployer) GetConfigMapByName(ctx context.Context, name string) (*corev1.ConfigMap, error) {
	return env.k8sClient.CoreV1().ConfigMaps(env.namespace).Get(ctx, name, metav1.GetOptions{})
}

func (env *K8sEnvDeployer) GetConfigMap(ctx context.Context, cfgMapName, cfgFileName string) ([]byte, error) {
	cfgMap, err := env.k8sClient.CoreV1().ConfigMaps(env.namespace).Get(ctx, cfgMapName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var tomlConfig interface{}
	for key, value := range cfgMap.Data {
		if key == cfgFileName {
			err := toml.Unmarshal([]byte(value), &tomlConfig)
			if err != nil {
				return nil, err
			}
			break
		}
	}

	tomlBytes, err := toml.Marshal(tomlConfig)
	if err != nil {
		fmt.Println("转换为TOML格式出错:", err)
		return nil, err
	}
	
	return tomlBytes, nil
}

func (env *K8sEnvDeployer) SetConfigMap(ctx context.Context, cfgMapName, cfgKey string, cfgValue []byte) error {
	configMapClient := env.k8sClient.CoreV1().ConfigMaps(env.namespace)
	cfgMap, err := configMapClient.Get(ctx, cfgMapName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	cfgMap.Data[cfgKey] = string(cfgValue)

	_, err = configMapClient.Update(ctx, cfgMap, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return err
}

func (env *K8sEnvDeployer) GetPod(ctx context.Context, podName string) (*corev1.Pod, error) {
	podClient := env.k8sClient.CoreV1().Pods(env.namespace)
	pod, err := podClient.Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pod, nil
}

func (env *K8sEnvDeployer) GetPodsByLabel(ctx context.Context, deployAppLabel ...string) ([]corev1.Pod, error) {
	podClient := env.k8sClient.CoreV1().Pods(env.namespace)
	podList, err := podClient.List(ctx, metav1.ListOptions{LabelSelector: "app in (" + strings.Join(deployAppLabel, ",") + ")"})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

func (env *K8sEnvDeployer) GetStatefulSet(ctx context.Context, name string) (*appv1.StatefulSet, error) {
	return env.k8sClient.AppsV1().StatefulSets(env.namespace).Get(ctx, name, metav1.GetOptions{})
}

func (env *K8sEnvDeployer) GetSvc(ctx context.Context, name string) (*corev1.Service, error) {
	return env.k8sClient.CoreV1().Services(env.namespace).Get(ctx, name, metav1.GetOptions{})
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

// ExecRemoteCmd execute remote server command in pod
func (env *K8sEnvDeployer) ExecRemoteCmd(ctx context.Context, podName string, cmd ...string) ([]byte, error) {
	stdOut := bytes.NewBuffer(nil)
	err := env.ExecRemoteCmdWithStream(ctx, podName, true, stdOut, nil, cmd...)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(stdOut)
}

// ExecRemoteCmd execute remote server command in pod
func (env *K8sEnvDeployer) ExecRemoteCmdWithStream(ctx context.Context, podName string, tty bool, stdOut io.Writer, stdIn io.Reader, cmd ...string) error {
	req := env.k8sClient.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace(env.namespace).SubResource("exec")
	option := &corev1.PodExecOptions{
		Command: cmd,
		Stdin:   false,
		Stdout:  false,
		Stderr:  true,
		TTY:     tty,
	}
	if stdOut != nil {
		option.Stdout = true
	}
	if stdIn != nil {
		option.Stdin = true
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(env.k8sCfg, "POST", req.URL())
	if err != nil {
		return err
	}

	streamOpts := remotecommand.StreamOptions{
		Stderr: os.Stderr,
		Tty:    tty,
	}
	if stdOut != nil {
		streamOpts.Stdout = stdOut
	}
	if stdIn != nil {
		streamOpts.Stdin = stdIn
	}
	return exec.StreamWithContext(ctx, streamOpts)
}

// ExecRemoteCmd execute remote server command in pod
func (env *K8sEnvDeployer) ExecRemoteCmdWithName(ctx context.Context, podName string, cmd ...string) ([]byte, error) {
	username := "ipfsman"
	password := "1"

	stdOut := bytes.NewBuffer(nil)
	stdIn := bytes.NewBuffer(nil)
	stdIn.WriteString(fmt.Sprintf("%s\n%s\n", username, password))

	err := env.ExecRemoteCmdWithStream(ctx, podName, true, stdOut, stdIn, cmd...)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(stdOut)
}

func (env *K8sEnvDeployer) WaitEndpointReady(ctx context.Context, endpoint types.Endpoint) error {
	for {
		time.Sleep(time.Second * 3)
		tCtx, cancel := context.WithTimeout(ctx, time.Second*5)
		_, err := env.dialCtx(tCtx, "tcp", string(endpoint))
		if err == nil {
			cancel()
			return err
		}
		cancel()
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

	errChan := make(chan error)
	go func() {
		err = fw.ForwardPorts()
		if err != nil {
			log.Errorf("forward port error %v", err)
			errChan <- err
		}
	}()

	go func() {
		<-ctx.Done()
		stopCh <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return "", errors.New("context cancel")
	case <-readyCh:
	case err := <-errChan:
		return "", err
	}

	return types.EndpointFromHostPort("127.0.0.1", freePort), nil
}

func (env *K8sEnvDeployer) Clean(ctx context.Context) error {
	return env.resourceMgr.Clean(ctx)
}

func UniqueId(testId string, retry int, instanceName string) string {
	return testId + strconv.Itoa(retry) + hex.EncodeToString(utils.Blake256([]byte(instanceName))[:4])
}
