package main

import (
	"fmt"
	clientset "github.com/Shaad7/bookstore-sample-controller/pkg/generated/clientset/versioned"
	informers "github.com/Shaad7/bookstore-sample-controller/pkg/generated/informers/externalversions"
	"github.com/Shaad7/bookstore-sample-controller/pkg/signals"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube/config")
	fmt.Println(kubeconfig)
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		panic(err)

	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err)

	}

	exampleClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	exampleInformerFactory := informers.NewSharedInformerFactory(exampleClient, time.Second*30)

	kubeInformerFactory = kubeInformerFactory
	exampleInformerFactory = exampleInformerFactory

	controller := NewController(kubeClient, exampleClient,
		kubeInformerFactory.Apps().V1().Deployments(),
		exampleInformerFactory.Gopher().V1alpha1().Bookstores())

	stopCh := signals.SetupSignalHandler()
	kubeInformerFactory.Start(stopCh)
	exampleInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		log.Println("Error running controller")
	}

}
