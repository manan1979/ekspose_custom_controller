package main 

import (
"k8s.io/client-go/kubernetes"
appListers "k8s.io/client-go/listers/apps/v1"
"k8s.io/client-go/tools/cache"
"k8s.io/client-go/util/workqueue"

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
      queue: workqueue.NewNameRateLimitingQueue(workqueue.DefaultControllerRateLimiter(),"ekspose"),
 }

 depInformer.Informer().AddEventHandler(
 cache.ResourceEventHandle{
              AddFunc: handleAd,
  DeleteFunc: handleDel,

},
 )
 return c 

}

func (c *controller) run(ch <-chan struct{}) {
if !cache.WaitForCacheSync(ch, c.depCacheSynced) {
   fmt.Print("waiting for cache to be synced\n")

}

   go wait.Unitl(c.worker, 1*time.Second, ch)

}

func (c *controller) worker() {

}

func handleAdd(obj interface{}) {

}


func handleDel(obj interface{}) {

}


