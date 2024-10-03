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

package framework

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os/exec"
	"time"

	"github.com/objectrocket/sensu-operator/pkg/client"
	"github.com/objectrocket/sensu-operator/pkg/generated/clientset/versioned"
	"github.com/objectrocket/sensu-operator/pkg/util/constants"
	"github.com/objectrocket/sensu-operator/pkg/util/k8sutil"
	"github.com/objectrocket/sensu-operator/pkg/util/retryutil"
	"github.com/objectrocket/sensu-operator/test/e2e/e2eutil"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var Global *Framework

type Framework struct {
	opImage    string
	KubeClient kubernetes.Interface
	CRClient   versioned.Interface
	Namespace  string
}

// Setup setups a test framework and points "Global" to it.
func setup() error {
	kubeconfig := flag.String("kubeconfig", "", "kube config path, e.g. $HOME/.kube/config")
	opImage := flag.String("operator-image", "", "operator image, e.g. sensu/sensu-operator:v0.0.1")
	ns := flag.String("namespace", "default", "e2e test namespace")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}
	cli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	Global = &Framework{
		KubeClient: cli,
		CRClient:   client.MustNew(config),
		Namespace:  *ns,
		opImage:    *opImage,
	}

	// Skip the sensu-operator deployment setup if the operator image was not specified
	if len(Global.opImage) == 0 {
		return nil
	}

	return Global.setup()
}

func teardown() error {
	// Skip the sensu-operator teardown if the operator image was not specified
	if len(Global.opImage) == 0 {
		return nil
	}

	err := Global.deleteOperatorCompletely("sensu-operator")
	if err != nil {
		return err
	}
	Global = nil
	logrus.Info("e2e teardown successfully")
	return nil
}

func (f *Framework) setup() error {
	err := f.SetupSensuOperator()
	if err != nil {
		return fmt.Errorf("failed to setup sensu operator: %v", err)
	}
	logrus.Info("sensu operator created successfully")

	logrus.Info("sensu setup successfully")
	return nil
}

func (f *Framework) SetupSensuOperator() error {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "sensu-operator",
			Labels: map[string]string{"name": "sensu-operator"},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				v1.Container{
					Name:  "sensu-operator",
					Image: f.opImage,
					// ImagePullPolicy: v1.PullAlways,
					Command: []string{"/usr/local/bin/sensu-operator"},
					Env: []v1.EnvVar{
						{
							Name:      constants.EnvOperatorPodNamespace,
							ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
						},
						{
							Name:      constants.EnvOperatorPodName,
							ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.name"}},
						},
					},
				},
			},
		},
	}

	p, err := k8sutil.CreateAndWaitPod(f.KubeClient, f.Namespace, pod, 60*time.Second)
	if err != nil {
		describePod(f.Namespace, "sensu-operator")
		return err
	}
	logrus.Infof("sensu operator pod is running on node (%s)", p.Spec.NodeName)

	return e2eutil.WaitUntilOperatorReady(f.KubeClient, f.Namespace, "sensu-operator")
}

func describePod(ns, name string) {
	// assuming `kubectl` installed on $PATH
	cmd := exec.Command("kubectl", "-n", ns, "describe", "pod", name)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run() // Just ignore the error...
	logrus.Infof("describing %s pod: %s", name, out.String())
}

func (f *Framework) DeleteSensuOperatorCompletely() error {
	return f.deleteOperatorCompletely("sensu-operator")
}

func (f *Framework) deleteOperatorCompletely(name string) error {
	ctx := context.Background()
	err := f.KubeClient.CoreV1().Pods(f.Namespace).Delete(ctx, name, *metav1.NewDeleteOptions(1))
	if err != nil {
		return err
	}
	// Grace period isn't exactly accurate. It took ~10s for operator pod to completely disappear.
	// We work around by increasing the wait time. Revisit this later.
	err = retryutil.Retry(5*time.Second, 6, func() (bool, error) {
		_, err := f.KubeClient.CoreV1().Pods(f.Namespace).Get(ctx, name, metav1.GetOptions{})
		if err == nil {
			return false, nil
		}
		if k8sutil.IsKubernetesResourceNotFoundError(err) {
			return true, nil
		}
		return false, err
	})
	if err != nil {
		return fmt.Errorf("fail to wait operator (%s) pod gone from API: %v", name, err)
	}
	return nil
}
