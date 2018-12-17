// Copyright 2017 The etcd-operator Authors
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

package e2eutil

import (
	"testing"
	"time"

	api "github.com/objectrocket/sensu-operator/pkg/apis/sensu/v1beta1"
	"github.com/objectrocket/sensu-operator/pkg/generated/clientset/versioned"
	"github.com/objectrocket/sensu-operator/pkg/util/k8sutil"
	"github.com/objectrocket/sensu-operator/pkg/util/retryutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/aws-sdk-go/service/s3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
)

type StorageCheckerOptions struct {
	S3Cli          *s3.S3
	S3Bucket       string
	DeletedFromAPI bool
}

func CreateCluster(t *testing.T, crClient versioned.Interface, namespace string, cl *api.SensuCluster) (*api.SensuCluster, error) {
	cl.Namespace = namespace
	res, err := crClient.SensuV1beta1().SensuClusters(namespace).Create(cl)
	if err != nil {
		return nil, err
	}
	t.Logf("creating sensu cluster: %s", res.Name)

	return res, nil
}

func UpdateCluster(crClient versioned.Interface, cl *api.SensuCluster, maxRetries int, updateFunc k8sutil.SensuClusterCRUpdateFunc) (*api.SensuCluster, error) {
	name := cl.Name
	namespace := cl.Namespace
	result := &api.SensuCluster{}
	err := retryutil.Retry(1*time.Second, maxRetries, func() (done bool, err error) {
		sensuCluster, err := crClient.SensuV1beta1().SensuClusters(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		updateFunc(sensuCluster)

		result, err = crClient.SensuV1beta1().SensuClusters(namespace).Update(sensuCluster)
		if err != nil {
			if apierrors.IsConflict(err) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	})
	return result, err
}

func DeleteCluster(t *testing.T, crClient versioned.Interface, kubeClient kubernetes.Interface, cl *api.SensuCluster) error {
	t.Logf("deleting sensu cluster: %v", cl.Name)
	err := crClient.SensuV1beta1().SensuClusters(cl.Namespace).Delete(cl.Name, nil)
	if err != nil {
		return err
	}
	return waitResourcesDeleted(t, kubeClient, cl)
}
