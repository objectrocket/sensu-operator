package controller

import (
	"fmt"

	api "github.com/objectrocket/sensu-operator/pkg/apis/objectrocket/v1beta1"
	sensu_client "github.com/objectrocket/sensu-operator/pkg/sensu_client"

	"k8s.io/client-go/tools/cache"
)

func (c *Controller) runCheckConfig() {
	for c.processNextCheckConfigItem() {
	}
}

func (c *Controller) processNextCheckConfigItem() bool {
	key, quit := c.checkConfigQueue.Get()
	if quit {
		return false
	}
	defer c.checkConfigQueue.Done(key)
	obj, exists, err := c.checkConfigIndexer.GetByKey(key.(string))
	if err != nil {
		if c.checkConfigQueue.NumRequeues(key) < c.Config.ProcessingRetries {
			c.checkConfigQueue.AddRateLimited(key)
			return true
		}
	} else {
		if !exists {
			c.onDeleteSensuCheckConfig(obj)
		} else {
			c.onUpdateSensuCheckConfig(obj)
		}
	}
	c.queue.Forget(key)
	return true
}

func (c *Controller) onUpdateSensuCheckConfig(newObj interface{}) {
	c.syncSensuCheckConfig(newObj.(*api.SensuCheckConfig))
}

func (c *Controller) onDeleteSensuCheckConfig(obj interface{}) {
	checkConfig, ok := obj.(*api.SensuCheckConfig)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			panic(fmt.Sprintf("unknown object from CheckConfig delete event: %#v", obj))
		}
		checkConfig, ok = tombstone.Obj.(*api.SensuCheckConfig)
		if !ok {
			panic(fmt.Sprintf("Tombstone contained object that is not a CheckConfig: %#v", obj))
		}
	}

	pt.start()
	sensuClient := sensu_client.New(checkConfig.Spec.SensuMetadata.Name, checkConfig.ObjectMeta.Namespace, checkConfig.Spec.SensuMetadata.Namespace)
	err := sensuClient.DeleteCheckConfig(checkConfig)
	if err != nil {
		c.logger.Warningf("fail to handle checkconfig delete event: %v", err)
	}
	pt.stop()
}

func (c *Controller) syncSensuCheckConfig(checkConfig *api.SensuCheckConfig) {
	pt.start()
	sensuClient := sensu_client.New(checkConfig.Spec.SensuMetadata.Name, checkConfig.ObjectMeta.Namespace, checkConfig.Spec.SensuMetadata.Namespace)
	err := sensuClient.UpdateCheckConfig(checkConfig)
	if err != nil {
		c.logger.Warningf("failed to handle checkconfig update event: %v", err)
	}
	copy := checkConfig.DeepCopy()
	copy.Status.Accepted = true
	if _, err = c.SensuCRCli.ObjectrocketV1beta1().SensuCheckConfigs(copy.GetNamespace()).Update(copy); err != nil {
		c.logger.Warningf("failed to update checkconfig's status during update event: %v", err)
	}
	pt.stop()
}
