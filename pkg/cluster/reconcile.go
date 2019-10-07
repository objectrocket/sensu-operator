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
	"context"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/objectrocket/sensu-operator/pkg/util/constants"
	"github.com/objectrocket/sensu-operator/pkg/util/etcdutil"
	"github.com/objectrocket/sensu-operator/pkg/util/k8sutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ErrLostQuorum indicates that the etcd cluster lost its quorum.
var ErrLostQuorum = errors.New("lost quorum")

// reconcile reconciles cluster current state to desired state specified by spec.
// - it tries to reconcile the cluster to desired size.
// - if the cluster needs for upgrade, it tries to upgrade old member one by one.
// TODO(tvoran): at least one of these statements is not true
func (c *Cluster) reconcile(pods []*v1.Pod) error {
	if c.statefulSet.Spec.Replicas == nil {
		c.logger.Infof("StatefulSet for cluster %s has nil Replicas.  Fetching new StatefulSet", c.name())
		set, err := c.config.KubeCli.AppsV1().StatefulSets(c.cluster.Namespace).Get(c.cluster.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("Failed to fetch new StatefulSet: %v", err)
		}
		c.statefulSet = set
		// Return here since reconcile will be called again and we can do this check again if necessary
		return nil
	}
	if c.cluster.Spec.Size != int(*c.statefulSet.Spec.Replicas) {
		set, err := c.config.KubeCli.AppsV1().StatefulSets(c.cluster.Namespace).Get(c.cluster.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("Error getting StatefulSet %s for size update: %v", c.statefulSet.GetName(), err)
		}
		var affectedMember int
		if int(*set.Spec.Replicas) < c.cluster.Spec.Size {
			*set.Spec.Replicas++
			affectedMember = int(*set.Spec.Replicas - 1)
			if err = c.addOneMember(affectedMember); err != nil {
				return err
			}
		} else {
			affectedMember = int(*set.Spec.Replicas - 1)
			*set.Spec.Replicas--
			if err = c.removeOneMember(affectedMember); err != nil {
				return err
			}
		}
		set, err = c.config.KubeCli.AppsV1().StatefulSets(c.cluster.Namespace).Update(set)
		if err != nil {
			return fmt.Errorf("Error updating StatefulSet %s size: %v", c.statefulSet.GetName(), err)
		}
		c.statefulSet = set
		c.status.Size = int(*set.Spec.Replicas)

		c.logger.Infof("Update StatefulSet %s size from %d to %d", c.statefulSet.GetName(), *c.statefulSet.Spec.Replicas, c.cluster.Spec.Size)
		return nil
	}
	var oldPod *v1.Pod
	oldPod = pickOneOldMember(pods, c.cluster.Spec.Version)
	if oldPod != nil {
		// This needs to be handled once the etcd cluster is either external or has multiple nodes
		c.logger.Warnf("Pod %s needs upgraded from version %s to %s", oldPod.GetName(), k8sutil.GetSensuVersion(oldPod), c.cluster.Spec.Version)
		return nil
	}
	return nil
}

func pickOneOldMember(pods []*v1.Pod, newVersion string) *v1.Pod {
	for _, pod := range pods {
		if k8sutil.GetSensuVersion(pod) == newVersion {
			continue
		}
		return pod
	}
	return nil
}

func (c *Cluster) addOneMember(ordinalID int) error {
	// c.status.SetScalingUpCondition(c.members.Size(), c.cluster.Spec.Size)
	m := &etcdutil.MemberConfig{
		Namespace:    c.cluster.Namespace,
		SecurePeer:   c.isSecurePeer(),
		SecureClient: c.isSecureClient(),
	}

	cfg := clientv3.Config{
		Endpoints:   c.ClientURLs(m),
		DialTimeout: constants.DefaultDialTimeout,
		TLS:         c.tlsConfig,
	}
	etcdcli, err := clientv3.New(cfg)
	if err != nil {
		return fmt.Errorf("add one member failed: creating etcd client failed %v", err)
	}
	defer etcdcli.Close()

	// newMember := c.newMember()
	ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultRequestTimeout)
	resp, err := etcdcli.MemberAdd(ctx, []string{c.PeerURL(m, ordinalID)})
	cancel()
	if err != nil {
		return fmt.Errorf("fail to add new member (%s): %v", c.memberName(ordinalID), err)
	}
	c.logger.Debugf("resp from memberadd was %+v", resp)
	// newMember.ID = resp.Member.ID
	// c.members.Add(newMember)

	// if err := c.createPod(c.members, newMember, "existing"); err != nil {
	// 	return fmt.Errorf("fail to create member's pod (%s): %v", newMember.Name, err)
	// }
	c.logger.Infof("added member (%s)", c.memberName(ordinalID))
	// _, err = c.eventsCli.Create(k8sutil.NewMemberAddEvent(c.memberName(ordinalID), c.cluster))
	// if err != nil {
	// 	c.logger.Errorf("failed to create new member add event: %v", err)
	// }
	return nil
}

func (c *Cluster) removeOneMember(ordinalID int) error {
	// TODO(tvoran): do the etcd setup in etcdutil, call etcdutil.RemoveMember
	// like in the old sensu-operator
	m := &etcdutil.MemberConfig{
		Namespace:    c.cluster.Namespace,
		SecurePeer:   c.isSecurePeer(),
		SecureClient: c.isSecureClient(),
	}

	cfg := clientv3.Config{
		Endpoints:   c.ClientURLs(m),
		DialTimeout: constants.DefaultDialTimeout,
		TLS:         c.tlsConfig,
	}
	etcdcli, err := clientv3.New(cfg)
	if err != nil {
		return fmt.Errorf("add one member failed: creating etcd client failed %v", err)
	}
	defer etcdcli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultRequestTimeout)
	defer cancel()

	mList, err := etcdcli.MemberList(ctx)
	if err != nil {
		return err
	}
	var id uint64
	memberName := fmt.Sprintf("%s-%d", c.name(), ordinalID)
	for _, m := range mList.Members {
		if m.Name == memberName {
			id = m.ID
		}
	}
	if id == 0 {
		return fmt.Errorf("Could not find %s in etcd member list", memberName)
	}

	resp, err := etcdcli.MemberRemove(ctx, id)
	// cancel()
	if err != nil {
		return fmt.Errorf("fail to remove member (%s): %v", c.memberName(ordinalID), err)
	}
	c.logger.Debugf("resp from MemberRemove was %+v", resp)

	if c.isPodPVEnabled() {
		err = c.removePVC(c.pvcName(ordinalID))
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Cluster) pvcName(ordinalID int) string {
	return fmt.Sprintf("etcd-data-%s-%d", c.name(), ordinalID)
}

func (c *Cluster) removePVC(pvcName string) error {
	err := c.config.KubeCli.Core().PersistentVolumeClaims(c.cluster.Namespace).Delete(pvcName, nil)
	if err != nil && !k8sutil.IsKubernetesResourceNotFoundError(err) {
		return fmt.Errorf("remove pvc (%s) failed: %v", pvcName, err)
	}
	return nil
}
