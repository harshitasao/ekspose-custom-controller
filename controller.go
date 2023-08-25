package main

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// make the controller struct
type controller struct {
	clientset      kubernetes.Interface
	depLister      appslisters.DeploymentLister
	depCacheSynced cache.InformerSynced
	queue          workqueue.RateLimitingInterface
}

func newController(clientset kubernetes.Interface, depInformer appsinformers.DeploymentInformer) *controller {
	c := &controller{
		clientset:      clientset,
		depLister:      depInformer.Lister(),
		depCacheSynced: depInformer.Informer().HasSynced,
		queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ekspose"),
	}
	depInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAdd,
			DeleteFunc: c.handleDel,
		},
	)
	return c
}

func (c *controller) run(ch <-chan struct{}) {
	fmt.Println("starting the controller")
	if !cache.WaitForCacheSync(ch, c.depCacheSynced) {
		fmt.Print("waiting for cache to sync")
	}

	go wait.Until(c.worker, 1*time.Second, ch)

	<-ch
}

func (c *controller) worker() {
	for c.processItem() {

	}
}

func (c *controller) processItem() bool {
	// here the bussiness logic is going to written
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	// if everything goes well and syncDeployment is successfull
	// then we want to remove the obj from the queue
	defer c.queue.Forget(item)

	// get the ns and name of obj we will use the cache
	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		fmt.Printf("getting key from cache %v", err.Error())
		return false
	}
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		fmt.Printf("splitting key into ns, name %v", err.Error())
		return false
	}
	err = c.syncDeployment(ns, name)
	if err != nil {
		// re-try
		fmt.Printf("sync deployment %v", err.Error())
		return false
	}
	return true
}

func (c *controller) syncDeployment(ns, name string) error {
	// create service

	dep, err := c.depLister.Deployments(ns).Get(name)
	if err != nil {
		fmt.Printf("getting deployment from lister %v", err.Error())
	}

	ctx := context.Background()

	// we have to modify this, and figure out that at which port
	// our deployment container is listening to
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dep.Name,
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Selector: depLabels(*dep),
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: 80,
				},
			},
		},
	}
	_, err = c.clientset.CoreV1().Services(ns).Create(ctx, &svc, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("creating service %v", err.Error())
	}

	// create ingress

	return nil
}

// func for getting labels from the backend pods
func depLabels(dep appsv1.Deployment) map[string]string {
	return dep.Spec.Template.Labels
}

// methods for controller
func (c *controller) handleAdd(obj interface{}) {
	fmt.Println("add was called")
	c.queue.Add(obj)
}

// methods for controller
func (c *controller) handleDel(obj interface{}) {
	fmt.Println("del was called")
	c.queue.Add(obj)
}
