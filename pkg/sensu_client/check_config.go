package client

import (
	"bytes"
	"errors"
	"reflect"
	"time"

	sensu_api_core_v2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"

	"github.com/objectrocket/sensu-operator/pkg/apis/objectrocket/v1beta1"
)

type fetchCheckResponse struct {
	checkConfig *types.CheckConfig
	err         error
}

// AddCheckConfig will add a new sensu checkconfig to the sensu server
func (s *SensuClient) AddCheckConfig(c *v1beta1.SensuCheckConfig) error {
	return s.ensureCheckConfig(c)
}

// UpdateCheckConfig will add a new sensu checkconfig to the sensu server
func (s *SensuClient) UpdateCheckConfig(c *v1beta1.SensuCheckConfig) error {
	return s.ensureCheckConfig(c)
}

// DeleteCheckConfig will delete an existing checkconfig from the sensu server
func (s *SensuClient) DeleteCheckConfig(c *v1beta1.SensuCheckConfig) error {
	c1 := make(chan error, 1)

	go func() {
		var err error
		if err = s.sensuCli.Client.DeleteCheck(c.ToSensuType()); err != nil {
			s.logger.Errorf("failed to delete checkconfig: %+v", err)
		}
		c1 <- err
	}()

	select {
	case err := <-c1:
		return err
	case <-time.After(10 * time.Second):
		s.logger.Warnf("timeout from sensu server after 10 seconds")
		return errors.New("timeout from sensu server after 10 seconds")
	}
}

func (s *SensuClient) ensureCheckConfig(c *v1beta1.SensuCheckConfig) error {
	var (
		check *types.CheckConfig
		err   error
	)

	if err := s.ensureCredentials(); err != nil {
		return err
	}

	c1 := make(chan fetchCheckResponse, 1)
	go func() {
		var (
			check *types.CheckConfig
			err   error
		)
		if check, err = s.sensuCli.Client.FetchCheck(c.Spec.SensuMetadata.Name); err != nil {
			s.logger.Warnf("failed to retrieve checkconfig name %s from namespace %s, err: %+v", c.Spec.SensuMetadata.Name, s.sensuCli.Config.Namespace(), err)
			// Assuming not found for now
			if err = s.sensuCli.Client.CreateCheck(c.ToSensuType()); err != nil {
				s.logger.Errorf("Failed to create new checkconfig: %s", err)
				c1 <- fetchCheckResponse{c.ToSensuType(), err}
			}
		}
		c1 <- fetchCheckResponse{check, nil}
	}()

	select {
	case response := <-c1:
		if response.err != nil {
			return response.err
		}
		check = response.checkConfig
	case <-time.After(10 * time.Second):
		s.logger.Warnf("timeout from sensu server after 10 seconds")
		return errors.New("timeout from sensu server after 10 seconds")
	}

	// Check to see if checkconfig needs updated?
	if !equal(check, c.ToSensuType()) {
		s.logger.Warnf("current checkconfig wasn't equal to new checkconfig, so updating...")
		c2 := make(chan error, 1)
		go func() {
			if err = s.sensuCli.Client.UpdateCheck(c.ToSensuType()); err != nil {
				s.logger.Errorf("Failed to update checkconfig: %s", err)
				c2 <- err
			}
			c2 <- nil
		}()

		select {
		case err = <-c2:
			return err
		case <-time.After(10 * time.Second):
			s.logger.Warnf("timeout from sensu server after 10 seconds")
			return errors.New("timeout from sensu server after 10 seconds")
		}
	}

	return nil
}

func equal(c1, c2 *sensu_api_core_v2.CheckConfig) bool {
	if c1 == nil || c2 == nil {
		return false
	}

	if c1.Command != c2.Command ||
		c1.HighFlapThreshold != c2.HighFlapThreshold ||
		c1.Interval != c2.Interval ||
		c1.LowFlapThreshold != c2.LowFlapThreshold ||
		c1.Publish != c2.Publish ||
		c1.ProxyEntityName != c2.ProxyEntityName ||
		c1.Stdin != c2.Stdin ||
		c1.Cron != c2.Cron ||
		c1.Ttl != c2.Ttl ||
		c1.Timeout != c2.Timeout ||
		c1.RoundRobin != c2.RoundRobin ||
		c1.OutputMetricFormat != c2.OutputMetricFormat ||
		!reflect.DeepEqual(c1.ObjectMeta, c2.ObjectMeta) {
		return false
	}

	for i, handler := range c1.Handlers {
		if handler != c2.Handlers[i] {
			return false
		}
	}

	for i, assets := range c1.RuntimeAssets {
		if assets != c2.RuntimeAssets[i] {
			return false
		}
	}

	for i, subscription := range c1.Subscriptions {
		if subscription != c2.Subscriptions[i] {
			return false
		}
	}

	if bytes.Compare(c1.ExtendedAttributes[:], c2.ExtendedAttributes[:]) != 0 {
		return false
	}

	if !c1.Subdue.Equal(c2.Subdue) {
		return false
	}

	if !c1.ProxyRequests.Equal(c2.ProxyRequests) {
		return false
	}

	for i, metricHandler := range c1.OutputMetricHandlers {
		if metricHandler != c2.OutputMetricHandlers[i] {
			return false
		}
	}

	for i, envVar := range c1.EnvVars {
		if envVar != c2.EnvVars[i] {
			return false
		}
	}

	return true
}
