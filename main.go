package main

import (
	"flag"
	"fmt"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "/home/harshitasao/.kube", "location to the kubeconfig file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("error %s in building config from flag \n", err.Error())

		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("error %s in getting inclusterconfig \n", err.Error())
		}

	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error in getting clientset %s\n", err.Error())
	}
	informer := informers.NewSharedInformerFactory(clientset, 10*time.Minute)
	fmt.Println(informer)
}
