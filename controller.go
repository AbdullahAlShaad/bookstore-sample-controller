package main

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"time"

	samplev1alpha1 "github.com/Shaad7/bookstore-sample-controller/pkg/apis/gopher/v1alpha1"
	clientset "github.com/Shaad7/bookstore-sample-controller/pkg/generated/clientset/versioned"
	samplescheme "github.com/Shaad7/bookstore-sample-controller/pkg/generated/clientset/versioned/scheme"
	informers "github.com/Shaad7/bookstore-sample-controller/pkg/generated/informers/externalversions/gopher/v1alpha1"
	listers "github.com/Shaad7/bookstore-sample-controller/pkg/generated/listers/gopher/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const controllerAgentName = "bookstore-controller"

const (
	SuccessSynced         = "Synced"
	ErrResourceExists     = "ErrResourceExists"
	MessageResourceExists = "Resource %q already exists and is not managed by Foo"
	MessageResourceSynced = "Bookstore synced successfully"
)

type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	sampleclientset clientset.Interface

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	bookstoreLister   listers.BookstoreLister
	bookstoreSynced   cache.InformerSynced

	workqueue workqueue.RateLimitingInterface

	recorder record.EventRecorder
}

func NewController(
	kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface,
	deploymentInformer appsinformers.DeploymentInformer,
	bookstoreInformer informers.BookstoreInformer) *Controller {

	utilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:     kubeclientset,
		sampleclientset:   sampleclientset,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		bookstoreLister:   bookstoreInformer.Lister(),
		bookstoreSynced:   bookstoreInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Foos"),
		recorder:          recorder,
	}

	klog.Info("Setting up event handlers")
	// To handle bookstore resource type changes
	bookstoreInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueBookstore,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueBookstore(new)
		},
	})
	// To handle deployment changes
	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*appsv1.Deployment)
			oldDepl := old.(*appsv1.Deployment)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {

				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})

	return controller
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("Starting controller")

	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.bookstoreSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {

	}
}
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but go # %v", obj))
			return nil
		}

		if err := c.syncHandler(key); err != nil {
			// could not reconcile
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%v', %v, requeueing", key, err.Error())
		}
		c.workqueue.Forget(obj)
		klog.Info("Reconciliation Successful")
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) syncHandler(key string) error {

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Invalid key %s", key))
		return nil
	}
	bookstore, err := c.bookstoreLister.Bookstores(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("Resource '%s' not found in workqueue", key))
			return nil
		}
		return err
	}
	deploymentName := bookstore.Spec.Name
	if deploymentName == "" {
		utilruntime.HandleError(fmt.Errorf("Deployment name is not specified for %s", key))
		return nil
	}
	deployement, err := c.deploymentsLister.Deployments(bookstore.Namespace).Get(deploymentName)

	if errors.IsNotFound(err) {
		deployement, err = c.kubeclientset.AppsV1().Deployments(bookstore.Namespace).Create(context.TODO(), newDeployment(bookstore), metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}

	if !metav1.IsControlledBy(deployement, bookstore) {
		msg := fmt.Sprintf(MessageResourceExists, deployement.Name)
		c.recorder.Event(bookstore, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf("%s", msg)
	}

	if bookstore.Spec.ReplicaCount != nil && *bookstore.Spec.ReplicaCount != *deployement.Spec.Replicas {
		klog.V(4).Infof("name : %s , Bookstore replicas : %d, deployment replicas : %d", name, *bookstore.Spec.ReplicaCount, *deployement.Spec.Replicas)
		deployement, err = c.kubeclientset.AppsV1().Deployments(bookstore.Namespace).Update(context.TODO(), newDeployment(bookstore), metav1.UpdateOptions{})
	}

	if err != nil {
		return err
	}

	err = c.updateBookstoreStatus(bookstore, deployement)
	if err != nil {
		return err
	}

	c.recorder.Event(bookstore, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) handleObject(obj interface{}) {

	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		klog.V(4).Infof("Recovered deleted object '%s' from tombstone")
	}
	klog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		if ownerRef.Kind != "Bookstore" {
			return
		}
		bookstore, err := c.bookstoreLister.Bookstores(object.GetNamespace()).Get(ownerRef.Name)

		if err != nil {
			klog.V(4).Infof("ignoring orphan object '%s' of bookstore '%s' ", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueBookstore(bookstore)
		return
	}
}

func (c *Controller) enqueueBookstore(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *Controller) updateBookstoreStatus(bookstore *samplev1alpha1.Bookstore, deployment *appsv1.Deployment) error {
	bookstoreCopy := bookstore.DeepCopy()
	bookstoreCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	_, err := c.sampleclientset.GopherV1alpha1().Bookstores(bookstore.Namespace).UpdateStatus(context.TODO(), bookstoreCopy, metav1.UpdateOptions{})
	return err
}

func newDeployment(bookstore *samplev1alpha1.Bookstore) *appsv1.Deployment {
	labels := map[string]string{
		"app":        "bookstore-api-server",
		"controller": bookstore.Name,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bookstore.Spec.Name,
			Namespace: bookstore.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(bookstore, samplev1alpha1.SchemeGroupVersion.WithKind("Bookstore")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: bookstore.Spec.ReplicaCount,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "bookstore-api-server",
							Image: "shaad7/bookstore-api-server:latest",
						},
					},
				},
			},
		},
	}
}
