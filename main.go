package main

import (
"flag"
"k8s.io/client-go/tools/clientcmd"
"fmt"                 
"k8s.io/client-go/kubernetes"
"k8s.io/client-go/rest"
"time"
"k8s.io/client-go/informers"
)


func main() {
   kubeconfig :=  flag.String("kubeconfig", "/home/ubuntu/.kube/config", "location to your kubeconfig  file")   
    

   config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)                                            
   if err !=nil {
       fmt.Printf("error %s building config from flag\n", err.Error())
   config ,err = rest.InClusterConfig()
     if err != nil {
       fmt.Printf("error %s getting in clusterconfig", err.Error())
     }
   }
 
   clientset, err :=  kubernetes.NewForConfig(config)
   if err != nil {
       fmt.Printf("error %s creating clientset\n" , err.Error())    
   }
   

   ch := make(chan struct{})
   informers :=    informers.NewSharedInformerFactory(clientset, 10 * time.Minute)
   

   c := newController(clientset,informers.Apps().V1().Deployments())
   informers.Start(ch)
   c.run(ch)
  fmt.Println(informers)


   }

   
