// informer.go
/*
package controller

import (
	"context"
	"time"

	api "github.com/objectrocket/sensu-operator/pkg/apis/objectrocket/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Informer is a struct that wraps the cache.Indexer and workqueue for a specific resource.
const CoreV1NodesPlural = "nodes"
var pt *panicTimer

func init() {
	pt = newPanicTimer(time.Minute, "unexpected long blocking (> 1 Minute) when handling cluster event")
}
type Informer struct {
	indexer   cache.Indexer
	queue     workqueue.RateLimitingInterface
	controller cache.Controller
}

// addInformer initializes an informer for a specific resource.
func (c *Controller) addInformer(namespace string, resourcePlural string, objType runtime.Object) {
	informer := Informer{
		queue: workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}

	source := cache.NewListWatchFromClient(
		c.Config.SensuCRCli.ObjectrocketV1beta1().RESTClient(),
		resourcePlural,
		namespace,
		fields.Everything(),
	)

	informer.indexer, informer.controller = cache.NewIndexerInformer(
		source,
		objType,
		c.ResyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAddFunc(resourcePlural),
			UpdateFunc: c.handleUpdateFunc(resourcePlural),
			DeleteFunc: c.handleDeleteFunc(resourcePlural),
		},
		cache.Indexers{},
	)

	c.informers[resourcePlural] = &informer
}

// handleAddFunc returns a function to handle add events.
func (c *Controller) handleAddFunc(resourcePlural string) func(interface{}) {
	return func(obj interface{}) {
		key, err := cache.MetaNamespaceKeyFunc(obj)
		if err == nil {
			c.informers[resourcePlural].queue.Add(key)
		}
	}
}

// handleUpdateFunc returns a function to handle update events.
func (c *Controller) handleUpdateFunc(resourcePlural string) func(interface{}, interface{}) {
	return func(old, new interface{}) {
		key, err := cache.MetaNamespaceKeyFunc(new)
		if err == nil {
			c.informers[resourcePlural].queue.Add(key)
		}
	}
}

// handleDeleteFunc returns a function to handle delete events.
func (c *Controller) handleDeleteFunc(resourcePlural string) func(interface{}) {
	return func(obj interface{}) {
		key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
		if err == nil {
			c.informers[resourcePlural].queue.Add(key)
		}
	}
}

// Start begins watching for events and processing them.
func (c *Controller) Start(ctx context.Context) {
	var ns string
	if c.Config.ClusterWide {
		ns = metav1.NamespaceAll
	} else {
		ns = c.Config.Namespace
	}

	c.addInformer(ns, api.SensuClusterResourcePlural, &api.SensuCluster{})
	// Add other informers similarly...

	c.startProcessing(ctx)
}

// startProcessing starts processing items in the queue.
func (c *Controller) startProcessing(ctx context.Context) {
	go wait.Until(c.run, time.Second, ctx.Done())
}

// run processes the next item in the queue.
func (c *Controller) run() {
	for c.processNextItem() {
	}
}

// processNextItem processes the next item in the queue.
func (c *Controller) processNextItem() bool {
	key, quit := c.informers[api.SensuClusterResourcePlural].queue.Get()
	if quit {
		return false
	}
	defer c.informers[api.SensuClusterResourcePlural].queue.Done(key)

	obj, exists, err := c.informers[api.SensuClusterResourcePlural].indexer.GetByKey(key.(string))
	if err != nil {
		c.handleErr(key, err)
		return true
	}

	if !exists {
		// Handle deletion logic if needed
	} else {
		c.onUpdateSensuClus(obj)
	}

	c.informers[api.SensuClusterResourcePlural].queue.Forget(key)
	return true
}

// handleErr handles errors when processing an item.
func (c *Controller) handleErr(key interface{}, err error) {
	if c.informers[api.SensuClusterResourcePlural].queue.NumRequeues(key) < c.Config.ProcessingRetries {
		c.informers[api.SensuClusterResourcePlural].queue.AddRateLimited(key)
		return
	}
	c.informers[api.SensuClusterResourcePlural].queue.Forget(key)
}

func (c *Controller) onUpdateSensuClus(obj interface{}) {
	cluster := obj.(*api.SensuCluster)
	// Handle update logic for SensuCluster
}
*/