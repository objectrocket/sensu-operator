package client

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"

	"github.com/sensu/sensu-go/types"
)

const (
	platformSensuNamespace = "platform"
)

type fetchEntityResponse struct {
	entity *types.Entity
	err    error
}

// AddNode will do nothing on a k8s node being added/updated/reconciled, for now
func (s *SensuClient) AddNode(node *corev1.Node) error {
	return s.ensureNode(node)
}

// UpdateNode will do nothing on a k8s node being added/updated/reconciled, for now
func (s *SensuClient) UpdateNode(node *corev1.Node) error {
	return s.ensureNode(node)
}

// DeleteNode will ensure that sensu entities associated with this k8s node are cleaned up
func (s *SensuClient) DeleteNode(nodeName string) error {
	if err := s.ensureCredentials(); err != nil {
		return errors.Wrap(err, "failed to ensure credentials for sensu client")
	}
	return s.ensureDeleteNode(nodeName)
}

// ensureNode left here for future use, as we potentially want to cleanup any dangling entities
func (s *SensuClient) ensureNode(node *corev1.Node) error {
	return nil
}

func (s *SensuClient) ensureDeleteNode(nodeName string) error {
	entity, err := s.fetchEntity(nodeName)
	if err != nil {
		return errors.Wrapf(err, "failed to find entity from node name %s", nodeName)
	}
	if entity == nil {
		return errors.New(fmt.Sprintf("failed to find entity from node name %s; empty entity", nodeName))
	}
	err = s.sensuCli.Client.DeleteEntity(entity.GetNamespace(), entity.GetName())
	if err != nil {
		s.logger.Warnf("failed to delete entity %+v from namespace %s, err: %+v", entity, entity.GetNamespace(), err)
		return errors.Wrapf(err, "failed to delete entity %+v from namespace %s", entity, entity.GetNamespace())
	}
	return nil
}

func (s *SensuClient) fetchEntity(nodeName string) (*types.Entity, error) {
	var (
		entity   *types.Entity
		err      error
	)
	c1 := make(chan fetchEntityResponse, 1)
	go func() {
		// Would love to use ListOptions{LabelSelector}, but that is an enterprise feature
		// if entities, err = s.sensuCli.Client.ListEntities(platformSensuNamespace, &client.ListOptions{
		// 	LabelSelector: labels.FormatLabels(map[string]string{"k8s_node": nodeName}),
		// })
		if entity, err = s.sensuCli.Client.FetchEntity(nodeName); err != nil {
			s.logger.Warnf("failed to retrieve entity %s from namespace %s, err: %+v", nodeName, platformSensuNamespace, err)
			c1 <- fetchEntityResponse{nil, errors.Wrapf(err, "failed to retrieve entity %s from namespace %s", nodeName, platformSensuNamespace)}
		}
		if entity != nil {
			s.logger.Debugf("found entity %s", entity.String())
			c1 <- fetchEntityResponse{entity, nil}
			return
		}
		s.logger.Debugf("nil entity was return for nodeName %s", nodeName)
		c1 <- fetchEntityResponse{nil, nil}
	}()

	select {
	case response := <-c1:
		if response.err != nil {
			return nil, response.err
		}
		entity = response.entity
	case <-time.After(s.timeout):
		s.logger.Warnf("timeout from sensu server after 10 seconds")
		return nil, errors.New("timeout from sensu server after 10 seconds")
	}
	return entity, nil
}