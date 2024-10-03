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

package cluster

import (
	"fmt"
	"strings"

	"github.com/objectrocket/sensu-operator/pkg/util/k8sutil"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (c *Cluster) upgradeOneMember(memberName string) error {
	//ctx := context.Background()
	c.status.SetUpgradingCondition(c.cluster.Spec.Version)

	ns := c.cluster.Namespace

	pod, err := c.config.KubeCli.CoreV1().Pods(ns).Get(memberName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("fail to get pod (%s): %v", memberName, err)
	}
	oldpod := pod.DeepCopy()

	c.logger.Infof("upgrading the sensu member %v from %s to %s", memberName, k8sutil.GetSensuVersion(pod), c.cluster.Spec.Version)
	pod.Spec.Containers[0].Image = k8sutil.ImageName(c.cluster.Spec.Repository, c.cluster.Spec.Version)
	k8sutil.SetSensuVersion(pod, c.cluster.Spec.Version)

	patchdata, err := k8sutil.CreatePatch(oldpod, pod, v1.Pod{})
	if err != nil {
		return fmt.Errorf("error creating patch: %v", err)
	}

	_, err = c.config.KubeCli.CoreV1().Pods(ns).Patch(pod.GetName(), types.StrategicMergePatchType, patchdata)
	if err != nil {
		return fmt.Errorf("fail to update the sensu member (%s): %v", memberName, err)
	}
	c.logger.Infof("finished upgrading the sensu member %v", memberName)
	_, err = c.eventsCli.Create(k8sutil.MemberUpgradedEvent(memberName, k8sutil.GetSensuVersion(oldpod), c.cluster.Spec.Version, c.cluster))
	if err != nil {
		c.logger.Errorf("failed to create member upgraded event: %v", err)
	}

	return nil
}

func (c *Cluster) upgradeStatefulSet() error {
	//ctx := context.Background()

	c.status.SetUpgradingCondition(c.cluster.Spec.Version)

	targetVersion := strings.Split(c.statefulSet.Spec.Template.Spec.Containers[0].Image, ":")[1]
	if targetVersion == c.cluster.Spec.Version {
		c.logger.Debugf("Waiting for StatefulSet rolling update")
		return nil
	}
	c.statefulSet.Spec.Template.Spec.Containers[0].Image = k8sutil.ImageName(c.cluster.Spec.Repository, c.cluster.Spec.Version)
	k8sutil.SetPodTemplateSensuVersion(&c.statefulSet.Spec.Template, c.cluster.Spec.Version)
	set, err := c.config.KubeCli.AppsV1().StatefulSets(c.cluster.Namespace).Update(c.statefulSet)
	if err != nil {
		return fmt.Errorf("failed to update sensu version in statefulset spec: %s", err)
	}
	c.statefulSet = set
	return nil
}
