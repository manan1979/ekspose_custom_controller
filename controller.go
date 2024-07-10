package main 

import (
"fmt"
"k8s.io/apimachinery/pkg/util/wait"
"time"
metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
appsinformers "k8s.io/client-go/informers/apps/v1"
"k8s.io/client-go/kubernetes"
appListers "k8s.io/client-go/listers/apps/v1"
"k8s.io/client-go/tools/cache"
"k8s.io/client-go/util/workqueue"
corev1 "k8s.io/api/core/v1"
"context"
)


type controller struct {
   clientset kubernetes.Interface
   depLister appListers.DeploymentLister
   depCacheSynced cache.InformerSynced
   queue workqueue.RateLimitingInterface
 

}


func newController(clientset kubernetes.Interface, depInformer appsinformers.DeploymentInformer) *controller {
 c := &controller{
      clientset: clientset,
      depLister: depInformer.Lister(),
      depCacheSynced: depInformer.Informer().HasSynced,
      queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(),"ekspose"),
 }

 depInformer.Informer().AddEventHandler(
 cache.ResourceEventHandlerFuncs{
              AddFunc: c.handleAdd,
              DeleteFunc: c.handleDel,

},
 )
 return c 

}


func (c *controller) run(ch <-chan struct{}) {
    fmt.Println("starting controller")
    if !cache.WaitForCacheSync(ch, c.depCacheSynced) {
   fmt.Print("waiting for cache to be synced\n")

}

   go wait.Until(c.worker, 1*time.Second, ch)
  
   <-ch 
}

func (c *controller) worker() {
  for c.processItem() {
   
  }
}

func (c *controller) processItem() bool {
    item, shutdown := c.queue.Get()
    if shutdown {
    return false
    }

    key, err := cache.MetaNamespaceKeyFunc(item)
    if err != nil {
    fmt.Printf("getting key from cache %s\n", err.Error())
    }

    ns, name, err := cache.SplitMetaNamespaceKey(key)
    if err != nil {
    fmt.Println("splitting key into namespace and name\n", err.Error())
    }
   
    err =  c.syncDeployment(ns,name)
    if err!= nil {
        //re-try
        fmt.Printf("syncing deployment %s\n",err.Error())
        return false 
    }
    return true


}

func (c *controller) syncDeployment(ns ,name string) error {
   ctx := context.Background()

  svc := corev1.Service{}
  _,err := c.clientset.CoreV1().Services(ns).Create(ctx, &svc, metav1.CreateOptions{} )
    if err != nil {
    fmt.Printf("creating service %s\n", err.Error())
    }
    return nil
  
}

func (c *controller) handleAdd(obj interface{}) {
 fmt.Println("add was called")
 c.queue.Add(obj)
}


func (c *controller) handleDel(obj interface{}) {
fmt.Println("delete was called")
}


