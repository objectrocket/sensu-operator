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
	c.logger.Debugf("key: %+v, obj: %+v, exists: %t, err: %+v from checkConfigIndexer.GetByKey", key, obj, exists, err)
	if err != nil {
		if c.checkConfigQueue.NumRequeues(key) < c.Config.ProcessingRetries {
			c.logger.Debugf("running checkConfigQueue.AddRateLimited(key) while managing checkconfigs")
			c.checkConfigQueue.AddRateLimited(key)
			return true
		}
	} else {
		if !exists {
			c.logger.Debugf("Calling onDeleteSensuCheckConfig with obj: %+v", obj)
			c.onDeleteSensuCheckConfig(obj)
		} else {
			c.logger.Debugf("Calling onUpdateSensuCheckConfig with obj: %+v", obj)
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
			// prevent panic on nil object.
			// TODO: why are these nil objects coming through?
			if obj == nil {
				return
			}
			panic(fmt.Sprintf("unknown object from CheckConfig delete event: %#v", obj))
		}
		checkConfig, ok = tombstone.Obj.(*api.SensuCheckConfig)
		if !ok {
			panic(fmt.Sprintf("Tombstone contained object that is not a CheckConfig: %#v", obj))
		}
	}

	// pt.start()
	sensuClient := sensu_client.New(checkConfig.Spec.SensuMetadata.Name, checkConfig.ObjectMeta.Namespace, checkConfig.Spec.SensuMetadata.Namespace)
	err := sensuClient.DeleteCheckConfig(checkConfig)
	if err != nil {
		c.logger.Warningf("fail to handle checkconfig delete event: %v", err)
	}
	// pt.stop()
}

func (c *Controller) syncSensuCheckConfig(checkConfig *api.SensuCheckConfig) {
	// pt.start()
	c.logger.Warnf("in syncSensuCheckConfig, about to update checkconfig within sensu cluster")
	sensuClient := sensu_client.New(checkConfig.Spec.SensuMetadata.Name, checkConfig.ObjectMeta.Namespace, checkConfig.Spec.SensuMetadata.Namespace)
	err := sensuClient.UpdateCheckConfig(checkConfig)
	c.logger.Warnf("in syncSensuCheckConfig, after update checkconfig in sensu cluster")
	if err != nil {
		c.logger.Warningf("failed to handle checkconfig update event: %v", err)
	}
	copy := checkConfig.DeepCopy()
	copy.Status.Accepted = true
	c.logger.Warnf("in syncSensuCheckConfig, about to update checkconfig status within k8s")
	if _, err = c.SensuCRCli.ObjectrocketV1beta1().SensuCheckConfigs(copy.GetNamespace()).Update(copy); err != nil {
		c.logger.Warningf("failed to update checkconfig's status during update event: %v", err)
	}
	c.logger.Warnf("in syncSensuCheckConfig, done updating checkconfig status within k8s")
	// pt.stop()
}
