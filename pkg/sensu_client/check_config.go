package client

import (
	"reflect"

	"github.com/sensu/sensu-go/types"

	"github.com/objectrocket/sensu-operator/pkg/apis/objectrocket/v1beta1"
)

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
	if err := s.sensuCli.Client.DeleteCheck(c.ToSensuType()); err != nil {
		s.logger.Errorf("failed to delete checkconfig: %+v", err)
		return err
	}
	return nil
}

func (s *SensuClient) ensureCheckConfig(c *v1beta1.SensuCheckConfig) error {
	var (
		check *types.CheckConfig
		err   error
	)

	if err := s.ensureCredentials(); err != nil {
		return err
	}

	if check, err = s.sensuCli.Client.FetchCheck(c.Spec.SensuMetadata.Name); err != nil {
		s.logger.Warnf("failed to retrieve checkconfig name %s from namespace %s, err: %+v", c.Spec.SensuMetadata.Name, s.sensuCli.Config.Namespace(), err)
		// Assuming not found for now
		if err = s.sensuCli.Client.CreateCheck(c.ToSensuType()); err != nil {
			s.logger.Errorf("Failed to create new checkconfig: %s", err)
			return err
		}
	}

	// Check to see if checkconfig needs updated?
	if !reflect.DeepEqual(check, c.ToSensuType()) {
		s.logger.Warnf("current checkconfig wasn't equal to new checkconfig, so updating...")
		if err = s.sensuCli.Client.UpdateCheck(c.ToSensuType()); err != nil {
			s.logger.Errorf("Failed to update checkconfig: %s", err)
			return err
		}
	}

	return nil
}
