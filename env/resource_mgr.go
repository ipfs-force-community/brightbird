package env

import (
	"context"
	"fmt"
	"strings"

	"github.com/hunjixin/brightbird/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type IResourceMgr interface {
	EnsureDatabase(string) error
	AssignDir() error
	Clean(context.Context) error
}

type ResourceMgr struct {
	tmpPath  string //shared root dir
	mysqlUrl string //Mysql connection string

	k8sClient *kubernetes.Clientset
	namespace string
	testId    string
}

func NewResourceMgr(k8sClient *kubernetes.Clientset, namespace, tmpPath, mysqlUrl, testId string) *ResourceMgr {
	return &ResourceMgr{
		k8sClient: k8sClient,
		tmpPath:   tmpPath,
		mysqlUrl:  mysqlUrl,
		namespace: namespace,
		testId:    testId,
	}
}

var _ IResourceMgr = (*ResourceMgr)(nil)

func (resourceMgr *ResourceMgr) EnsureDatabase(dsn string) error {
	if !strings.Contains(dsn, resourceMgr.testId) {
		//not a random database
		return nil
	}

	return utils.CreateDatabase(dsn)
}

func (resourceMgr *ResourceMgr) AssignDir() error {
	panic("not implemented") // TODO: assign random directory or use exist directory
}

func (resourceMgr *ResourceMgr) Clean(ctx context.Context) error {
	err := resourceMgr.k8sClient.AppsV1().Deployments(resourceMgr.namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: "testid=" + resourceMgr.testId})
	if err != nil {
		log.Errorf("clean deployment failed %s", err)
	}
	log.Debug("celan deployment success")

	err = resourceMgr.k8sClient.AppsV1().StatefulSets(resourceMgr.namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: "testid=" + resourceMgr.testId})
	if err != nil {
		log.Errorf("clean statefuleset failed %s", err)
	}
	log.Debug("celan statefulset success")

	err = resourceMgr.k8sClient.CoreV1().Pods(resourceMgr.namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: "testid=" + resourceMgr.testId})
	if err != nil {
		log.Errorf("clean pod failed %s", err)
	}
	log.Debug("celan pods success")

	services, err := resourceMgr.k8sClient.CoreV1().Services(resourceMgr.namespace).List(ctx, metav1.ListOptions{LabelSelector: "testid=" + resourceMgr.testId})
	if err != nil {
		log.Errorf("clean service failed %s", err)
	}
	log.Debug("celan service success")

	err = resourceMgr.k8sClient.CoreV1().ConfigMaps(resourceMgr.namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: "testid=" + resourceMgr.testId})
	if err != nil {
		log.Errorf("clean configmap failed %s", err)
	}
	log.Debug("celan configmap success")

	err = resourceMgr.k8sClient.CoreV1().PersistentVolumeClaims(resourceMgr.namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: "testid=" + resourceMgr.testId})
	if err != nil {
		log.Errorf("clean pvc failed %s", err)
	}
	log.Debug("celan pv success")

	err = resourceMgr.k8sClient.CoreV1().PersistentVolumes().DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: "testid=" + resourceMgr.testId})
	if err != nil {
		log.Errorf("clean pv failed %s", err)
	}
	log.Debug("celan pv success")

	for _, svc := range services.Items {
		err := resourceMgr.k8sClient.CoreV1().Services(resourceMgr.namespace).Delete(ctx, svc.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Errorf("clean service failed %s", err)
		}
	}
	log.Debug("celan services success")

	databases, err := utils.ListDatabase(resourceMgr.mysqlUrl)
	if err != nil {
		log.Errorf("list databases failed %s", err)
	}

	for _, db := range databases {
		if strings.Contains(db, resourceMgr.testId) {
			err = utils.DropDatabase(fmt.Sprintf(resourceMgr.mysqlUrl, db))
			if err != nil {
				log.Errorf("drop %s databases failed %s", db, err)
			}
		}
	}
	return nil
}
