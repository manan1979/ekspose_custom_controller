package main 

import (
"fmt"
"k8s.io/apimachinery/pkg/util/wait"
"time"

netv1 "k8s.io/api/networking/v1"
appsv1 "k8s.io/api/apps/v1"
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
    defer c.queue.Forget(item)
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

   dep, err := c.depLister.Deployments(ns).Get(name)
      if err != nil {
       fmt.Printf("getting deployment from lister %s\n", err.Error())
   }

  svc := corev1.Service{
     ObjectMeta : metav1.ObjectMeta{
         Name : dep.Name,
         Namespace : ns,
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
  s, err :=  c.clientset.CoreV1().Services(ns).Create(ctx, &svc, metav1.CreateOptions{} )
    if err != nil {
        fmt.Printf("creating service %s\n", err.Error())
    }

    return createIngress(ctx, c.clientset, s)
}

func createIngress(ctx context.Context ,client kubernetes.Interface, svc *corev1.Service) error {  
    pathType := "Prefix"
    ingress := netv1.Ingress{
       ObjectMeta: metav1.ObjectMeta{
         Name: svc.Name,
         Namespace: svc.Namespace,
         Annotations: map[string]string{
           "nginx.ingress.kubernetes.io/rewrite-target": "/",
         },
       },
       Spec: netv1.IngressSpec{
         Rules:  []netv1.IngressRule{
          netv1.IngressRule{
           IngressRuleValue: netv1.IngressRuleValue{
              HTTP: &netv1.HTTPIngressRuleValue{
                  Paths: []netv1.HTTPIngressPath{
                     netv1.HTTPIngressPath{
                        Path:  fmt.Sprintf("/%s", svc.Name),
                        PathType: (*netv1.PathType)(&pathType),
                        Backend: netv1.IngressBackend{
                           Service: &netv1.IngressServiceBackend{
                             Name: svc.Name,
                             Port: netv1.ServiceBackendPort{
                                   Number: 80,
                             },
                           },                                        
                        },
                     },
                  },
              },
           },
           },
         },
       },
    }
    _,err := client.NetworkingV1().Ingresses(svc.Namespace).Create(ctx, &ingress, metav1.CreateOptions{})
    return err
}

func depLabels(dep  appsv1.Deployment)  map[string] string  {
    return dep.Spec.Template.Labels
}

func (c *controller) handleAdd(obj interface{}) {
 fmt.Println("add was called")
 c.queue.Add(obj)
}


func (c *controller) handleDel(obj interface{}) {
fmt.Println("delete was called")
}


