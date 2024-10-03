// Copyright 2016 The etcd-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"context"
	"fmt"
	"time"

	api "github.com/objectrocket/sensu-operator/pkg/apis/objectrocket/v1beta1"
	"github.com/objectrocket/sensu-operator/pkg/cluster"
	"github.com/objectrocket/sensu-operator/pkg/generated/clientset/versioned"
	"github.com/objectrocket/sensu-operator/pkg/util/k8sutil"
	sensucli "github.com/sensu/sensu-go/cli"
	"github.com/sirupsen/logrus"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1" // Import for CustomResourceValidation
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var initRetryWaitTime = 30 * time.Second

// Event is the cluster event that pairs the event type (add/update/delete)
// with the sensu cluster object, and is passed through the controller
type Event struct {
	Type   kwatch.EventType
	Object *api.SensuCluster
}

type getByKey interface {
	GetByKey(key string) (item interface{}, exists bool, err error)
}

type hasSynced interface {
	Run(stopCh <-chan struct{})
	HasSynced() bool
}

type rateLimitedQueue interface {
	Add(item interface{})
	Done(item interface{})
	ShutDown()
	Get() (item interface{}, shutdown bool)
	Forget(item interface{})
	NumRequeues(item interface{}) int
	AddRateLimited(item interface{})
}

// Informer is a kubernetes informer that satisfies the included 3 interfaces
type Informer struct {
	indexer    getByKey
	controller hasSynced
	queue      rateLimitedQueue
}

// Controller is the sensu controller that handles all informers/clusters/finalizers
// for all of the custom resources in the operator
type Controller struct {
	logger *logrus.Entry
	Config
	informers  map[string]*Informer
	clusters   map[string]*cluster.Cluster
	finalizers map[string]cache.Indexer
}

// Config is the configuration for the sensu controller
type Config struct {
	Namespace         string
	ClusterWide       bool
	ServiceAccount    string
	KubeCli           kubernetes.Interface
	KubeExtCli        apiextensionsclient.Interface
	SensuCRCli        versioned.Interface
	CreateCRD         bool
	WorkerThreads     int
	ProcessingRetries int
	ResyncPeriod      time.Duration
	LogLevel          logrus.Level
}

func clientForCluster(name string) (*sensucli.SensuCli, error) {
	return nil, nil
}

// New returns a new sensu controller
func New(cfg Config) *Controller {
	logrus.SetLevel(cfg.LogLevel)
	return &Controller{
		logger:     logrus.WithField("pkg", "controller"),
		informers:  make(map[string]*Informer),
		Config:     cfg,
		clusters:   make(map[string]*cluster.Cluster),
		finalizers: make(map[string]cache.Indexer),
	}
}

// handleClusterEvent returns true if cluster is ignored (not managed) by this instance.
func (c *Controller) handleClusterEvent(event *Event) (bool, error) {
	ctx := context.Background()
	var err error
	clus := event.Object

	if !c.managed(clus) {
		return true, nil
	}

	if clus.Status.IsFailed() {
		clustersFailed.Inc()
		if event.Type == kwatch.Deleted {
			delete(c.clusters, clus.Name)
			return false, nil
		}
		return false, fmt.Errorf("ignore failed cluster (%s). Please delete its CR", clus.Name)
	}

	clus.SetDefaults()

	if err := clus.Spec.Validate(); err != nil {
		return false, fmt.Errorf("invalid cluster spec. please fix the following problem with the cluster spec: %v", err)
	}

	switch event.Type {
	case kwatch.Added:
		if _, ok := c.clusters[clus.Name]; ok {
			return false, fmt.Errorf("unsafe state. cluster (%s) was created before but we received event (%s)", clus.Name, event.Type)
		}

		c.clusters[clus.Name] = cluster.New(c.makeClusterConfig(), clus)
		clustersCreated.Inc()
		clustersTotal.Inc()

	case kwatch.Modified:
		if _, ok := c.clusters[clus.Name]; !ok {
			return false, fmt.Errorf("unsafe state. cluster (%s) was never created but we received event (%s)", clus.Name, event.Type)
		}
		c.clusters[clus.Name].Update(clus)
		clustersModified.Inc()

	case kwatch.Deleted:
		if _, ok := c.clusters[clus.Name]; !ok {
			return false, fmt.Errorf("unsafe state. cluster (%s) was never created but we received event (%s)", clus.Name, event.Type)
		}
		c.clusters[clus.Name].Delete()
		deletionPolicy := v1.DeletePropagationBackground
		err = c.KubeCli.AppsV1().StatefulSets(clus.GetNamespace()).Delete(ctx, clus.GetName(), v1.DeleteOptions{
			PropagationPolicy: &deletionPolicy,
		})
		if err != nil {
			if k8sutil.IsKubernetesResourceNotFoundError(err) {
				c.logger.Warnf("StatefulSet %s does not exist", clus.GetName())

			} else {
				c.logger.Errorf("Failed to delete StatefulSet %s", clus.GetName())
			}
		}
		c.logger.Infof("Deleted cluster %s", clus.GetName())
		delete(c.clusters, clus.Name)
		clustersDeleted.Inc()
		clustersTotal.Dec()
	}
	return false, nil
}

func (c *Controller) makeClusterConfig() cluster.Config {
	return cluster.Config{
		ServiceAccount: c.Config.ServiceAccount,
		KubeCli:        c.Config.KubeCli,
		SensuCRCli:     c.Config.SensuCRCli,
	}
}
func (c *Controller) initCRD() (err error) {
	crds := []struct {
		name       string
		kind       string
		plural     string
		shortName  string
		validation *apiextensionsv1.CustomResourceValidation
	}{
		{api.SensuClusterCRDName, api.SensuClusterResourceKind, api.SensuClusterResourcePlural, "sensu", nil},
		{api.SensuAssetCRDName, api.SensuAssetResourceKind, api.SensuAssetResourcePlural, "sensuasset", api.SensuAsset{}.GetCustomResourceValidation()},
		{api.SensuCheckConfigCRDName, api.SensuCheckConfigResourceKind, api.SensuCheckConfigResourcePlural, "sensucheckconfig", api.SensuCheckConfig{}.GetCustomResourceValidation()},
		{api.SensuHandlerCRDName, api.SensuHandlerResourceKind, api.SensuHandlerResourcePlural, "sensuhandler", api.SensuHandler{}.GetCustomResourceValidation()},
		{api.SensuEventFilterCRDName, api.SensuEventFilterResourceKind, api.SensuEventFilterResourcePlural, "sensueventfilter", api.SensuEventFilter{}.GetCustomResourceValidation()},
		// Add other CRDs here...
	}

	for _, crd := range crds {
		// Call CreateCRD with the correct number of parameters
		if err = k8sutil.CreateCRD(c.KubeExtCli, crd.name, crd.kind, crd.plural, crd.shortName, crd.validation); err != nil {
			c.logger.Errorf("Failed to create %s CRD: %v", crd.name, err)
			return fmt.Errorf("failed to create %s CRD: %v", crd.name, err)
		}
		if err = k8sutil.WaitCRDReady(c.KubeExtCli, crd.name); err != nil {
			c.logger.Errorf("Failed to wait for %s CRD to become ready: %v", crd.name, err)
			return fmt.Errorf("failed to wait for %s CRD: %v", crd.name, err)
		}
	}
	return nil
}

func (c *Controller) clusterExists(clusterName string) (ok bool) {
	_, ok = c.clusters[clusterName]
	return
}
