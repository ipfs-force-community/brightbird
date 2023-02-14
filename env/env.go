package env

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/filecoin-project/venus-auth/jwtclient"
	"google.golang.org/appengine"
	"io"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	yaml_k8s "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type CloseFunc func() error

func JoinCloser(c1, c2 CloseFunc) CloseFunc {
	return func() error {
		mErr := appengine.MultiError{}
		if err := c1(); err != nil {
			mErr = append(mErr, err)
		}
		if err := c2(); err != nil {
			mErr = append(mErr, err)
		}
		if len(mErr) == 0 {
			return nil
		}
		return mErr
	}
}

type EnvController struct {
	k8sClient *kubernetes.Clientset
	namespace string
	hostIP    string
	testId    string
	cfg       *rest.Config
	debug     bool
}

func NewEnvController(namespace string, testId string, debug bool) (*EnvController, error) {
	var config *rest.Config
	var err error
	if debug {
		var kubeConfig *string
		if home := homedir.HomeDir(); home != "" {
			kubeConfig = flag.String("kubeConfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeConfig file")
		} else {
			kubeConfig = flag.String("kubeConfig", "", "absolute path to the kubeConfig file")
		}
		flag.Parse()

		// use the current context in kubeConfig
		config, err = clientcmd.BuildConfigFromFlags("", *kubeConfig)
		if err != nil {
			return nil, err
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
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

	return &EnvController{
		cfg:       config,
		k8sClient: k8sClient,
		namespace: namespace,
		hostIP:    url.Hostname(),
		testId:    testId,
		debug:     debug,
	}, nil
}

func (env *EnvController) RunVenusAuth(ctx context.Context, scriptPath string) (jwtclient.IAuthClient, CloseFunc, error) {
	_, closer, err := env.runDeployment(ctx, filepath.Join(scriptPath, "venus-auth-deployment.yaml"))
	if err != nil {
		return nil, nil, err
	}

	serviceCfgName := "venus-auth-service.yaml"
	if env.debug {
		serviceCfgName = "venus-auth-service-nodeport.yaml"
	}
	_, endpoint, closer2, err := env.runService(ctx, filepath.Join(scriptPath, serviceCfgName))
	if err != nil {
		closer()
		return nil, nil, err
	}

	closer = JoinCloser(closer, closer2)
	authClient, err := jwtclient.NewAuthClient("http://" + endpoint)
	if err != nil {
		closer()
		return nil, nil, err
	}
	return authClient, closer, nil
}

func (env *EnvController) runDeployment(ctx context.Context, deploymentCfgPath string) (*appv1.Deployment, CloseFunc, error) {
	data, err := env.readAndReplace(deploymentCfgPath)
	if err != nil {
		return nil, nil, err
	}

	fmt.Println(string(data))
	deployment := &appv1.Deployment{}
	err = yaml_k8s.Unmarshal(data, deployment)
	if err != nil {
		return nil, nil, err
	}

	name := deployment.Name
	deploymentClient := env.k8sClient.AppsV1().Deployments(env.namespace)
	fmt.Println("Creating deployment...")
	result, err := deploymentClient.Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, nil, err
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	closeFunc := func() error {
		return deploymentClient.Delete(ctx, name, metav1.DeleteOptions{})
	}

	for {
		select {
		case <-ctx.Done():
			closeFunc()
			return nil, nil, fmt.Errorf("context cancel when deploy %s", name)
		default:
			dep, err := deploymentClient.Get(ctx, deployment.GetName(), metav1.GetOptions{})
			if err != nil {
				closeFunc()
				return nil, nil, err
			}

			if dep.Status.ReadyReplicas == *deployment.Spec.Replicas {
				return dep, closeFunc, nil
			}
			time.Sleep(time.Second * 5)
		}
	}
}

func (env *EnvController) runService(ctx context.Context, serviceCfgPath string) (*corev1.Service, string, CloseFunc, error) {
	data, err := env.readAndReplace(serviceCfgPath)
	if err != nil {
		return nil, "", nil, err
	}

	fmt.Println(string(data))
	serviceCfg := &corev1.Service{}
	err = yaml_k8s.Unmarshal(data, serviceCfg)
	if err != nil {
		return nil, "", nil, err
	}

	name := serviceCfg.Name
	serviceClient := env.k8sClient.CoreV1().Services(env.namespace)
	fmt.Println("Creating service...")
	result, err := serviceClient.Create(ctx, serviceCfg, metav1.CreateOptions{})
	if err != nil {
		return nil, "", nil, err
	}
	fmt.Printf("Created service %q.\n", result.GetObjectMeta().GetName())

	closeFunc := func() error {
		return serviceClient.Delete(ctx, name, metav1.DeleteOptions{})
	}

	for {
		select {
		case <-ctx.Done():
			closeFunc()
			return nil, "", nil, fmt.Errorf("context cancel when deploy %s", name)
		default:
			service, err := serviceClient.Get(ctx, serviceCfg.GetName(), metav1.GetOptions{})
			if err != nil {
				closeFunc()
				return nil, "", nil, err
			}
			endpoint := ""
			if service.Spec.Type == corev1.ServiceTypeClusterIP {
				//todo check service is work
				if len(service.Spec.ClusterIP) > 0 {
					//take first
					return service, fmt.Sprintf("%s:%s", service.Spec.ClusterIP, strconv.Itoa(int(service.Spec.Ports[0].Port))), closeFunc, nil
				}
				closeFunc()
				return nil, "", nil, fmt.Errorf("unable to get cluser ip for %s", name)

			} else if service.Spec.Type == corev1.ServiceTypeNodePort {
				endpoint = fmt.Sprintf("%s:%s", env.hostIP, strconv.Itoa(int(service.Spec.Ports[0].NodePort)))
			} else {
				closeFunc()
				return nil, "", nil, fmt.Errorf("unable service type %s(%s)", name, service.Spec.Type)
			}

			var d net.Dialer
			ctx, _ = context.WithTimeout(ctx, time.Second*5)
			_, err = d.DialContext(ctx, "tcp", endpoint)
			if err == nil {
				return service, endpoint, closeFunc, nil
			}
			continue
		}
	}
}

func (env *EnvController) readAndReplace(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return bytes.ReplaceAll(data, []byte("{testid}"), []byte(env.testId)), nil
}

func (env *EnvController) readSmallFielInPod(ctx context.Context, deployAppLabel string, path string) ([]byte, error) {
	podClient := env.k8sClient.CoreV1().Pods(env.namespace)
	pods, err := podClient.List(ctx, metav1.ListOptions{LabelSelector: "app=" + deployAppLabel})
	if err != nil {
		return nil, err
	}

	cmd := []string{
		"ssh",
		"-c",
		"cat ",
		path,
	}
	req := env.k8sClient.CoreV1().RESTClient().Post().Resource("pods").Name(pods.Items[0].GetName()).
		Namespace(env.namespace).SubResource("exec")
	option := &corev1.PodExecOptions{
		Command: cmd,
		Stdin:   false,
		Stdout:  true,
		Stderr:  false,
		TTY:     false,
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(env.cfg, "POST", req.URL())
	if err != nil {
		return nil, err
	}
	stdOut := bytes.NewBuffer(nil)
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: stdOut,
	})
	return io.ReadAll(stdOut)
}
