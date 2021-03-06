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

package absfactory

import (
	"fmt"

	api "github.com/objectrocket/sensu-operator/pkg/apis/objectrocket/v1beta1"

	"github.com/Azure/azure-sdk-for-go/storage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ABSClient is a wrapper of ABS client that provides cleanup functionality.
type ABSClient struct {
	ABS *storage.BlobStorageClient
}

// NewClientFromSecret returns a ABS client based on given k8s secret containing azure credentials.
func NewClientFromSecret(kubecli kubernetes.Interface, namespace, absSecret string) (w *ABSClient, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("new ABS client failed: %v", err)
		}
	}()

	se, err := kubecli.CoreV1().Secrets(namespace).Get(absSecret, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get k8s secret: %v", err)
	}

	storageAccount := se.Data[api.AzureSecretStorageAccount]
	storageKey := se.Data[api.AzureSecretStorageKey]

	bc, err := storage.NewBasicClient(
		string(storageAccount),
		string(storageKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure storage client: %v", err)
	}

	abs := bc.GetBlobService()
	return &ABSClient{ABS: &abs}, nil
}
