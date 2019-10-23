package controller

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/tools/cache"

	sensu_client "github.com/objectrocket/sensu-operator/pkg/sensu_client"
)

const (
	platformSensuClusterName    = "sensu"
	platformSensuNamespace      = "platform"
	platformKubernetesNamespace = "sensu"
)

func (c *Controller) onUpdateNode(newObj interface{}) {
	c.syncNode(newObj.(*corev1.Node))
}

func (c *Controller) onDeleteNode(obj interface{}) {
	node, ok := obj.(*corev1.Node)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			// prevent panic on nil object/such as actual deletion
			if obj == nil {
				return
			}
			panic(fmt.Sprintf("unknown object from Node delete event: %#v", obj))
		}
		node, ok = tombstone.Obj.(*corev1.Node)
		if !ok {
			panic(fmt.Sprintf("Tombstone contained object that is not a Node: %#v", obj))
		}
	}

	if c.clusterExists(platformSensuClusterName) {
		sensuClient := sensu_client.New(platformSensuClusterName, platformKubernetesNamespace, platformSensuNamespace)
		err := sensuClient.DeleteNode(node)
		if err != nil {
			c.logger.Warningf("failed to handle node delete event: %v", err)
			return
		}
	}
}

func (c *Controller) syncNode(*corev1.Node) {
}
