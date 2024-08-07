package main

import (
	"flag"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	//Load kubeconfig- flag is a function which will parse command line argument (name,default value,description)
	kubeconfig := flag.String("kubeconfig", "/path/to/kubeconfig", "path to the kubeconfig file")
	flag.Parse()

	//fetch kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	//create clientset object from kubeconfig
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	//shared factory informers lets you watch multiple CRs parallely
	factory := informers.NewSharedInformerFactory(clientset, time.Second*30)

	//creating pod informer
	podInformer := factory.Core().V1().Pods().Informer()

	//Event handler has 3 lifecycle, which get triggered based on the event
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
            pod := obj.(*corev1.Pod)
            fmt.Printf("Pod added: %s/%s\n", pod.Namespace, pod.Name)
        },
        UpdateFunc: func(oldObj, newObj interface{}) {
            oldPod := oldObj.(*corev1.Pod)
            newPod := newObj.(*corev1.Pod)
            fmt.Printf("Pod updated: %s/%s\n", oldPod.Namespace, oldPod.Name)
            fmt.Printf("New Pod: %s/%s\n", newPod.Namespace, newPod.Name)
        },
        DeleteFunc: func(obj interface{}) {
            pod := obj.(*corev1.Pod)
            fmt.Printf("Pod deleted: %s/%s\n", pod.Namespace, pod.Name)
        },
	})

	//channel is created to stop the informer gracefully
	stopCh := make(chan struct{})
	defer close(stopCh)
	factory.Start(stopCh)

	//informer will watch for the any changes of CRd and update in Cache, which then consumed by event listeners
	if ok := cache.WaitForCacheSync(stopCh, podInformer.HasSynced); !ok {
        fmt.Println("Failed to wait for caches to sync")
        return
    }

	//For infinitely running the informer
	<-stopCh
}

/*
PRE-REQUISITES
mkdir kube-informer
cd kube-informer
go mod init kube-informer

go get k8s.io/client-go@latest
go get k8s.io/api@latest
go get k8s.io/apimachinery@latest
*/
