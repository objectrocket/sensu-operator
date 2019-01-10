package client

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"

	"github.com/sirupsen/logrus"

	"github.com/objectrocket/sensu-operator/pkg/util/k8sutil"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/client/config/basic"
	"github.com/sensu/sensu-go/types"
)

// SensuClient is the sensu client used to interact with a sensu server within
// a kubernetes cluster, within a certain k8s namespace
//
// NOTE: ** this is tied to a sensu namespace, so a new SensuClient is needed
//   between sensu namespaces **
//
// TODO: add a factory method that returns a sensuclient, and allows easy switching
// of sensu namespaces
type SensuClient struct {
	logger *logrus.Entry

	clusterName string
	namespace   string

	sensuCli *cli.SensuCli
}

// New will return a new SensuClient tied to a specific cluster within a k8s
// namespace, and tied to a specific sensu namespace.
func New(clusterName, namespace string, sensuNamespace string) *SensuClient {
	sClient := &SensuClient{
		logger:      logrus.WithField("pkg", "sensu_client").WithField("cluster-name", clusterName),
		clusterName: clusterName,
		namespace:   namespace,
	}

	conf := basic.Config{
		Cluster: basic.Cluster{
			APIUrl:  fmt.Sprintf("http://%s:8080", sClient.makeFullyQualifiedSensuClientURL()),
			Edition: "enterprise",
		},
		Profile: basic.Profile{
			Format:    "json",
			Namespace: sensuNamespace,
		},
	}

	sensuCliClient := client.New(&conf)
	logger := logrus.WithFields(logrus.Fields{
		"component": "cli-client",
	})

	sClient.sensuCli = &cli.SensuCli{
		Client: sensuCliClient,
		Config: &conf,
		Logger: logger,
	}

	return sClient
}

func (s *SensuClient) makeFullyQualifiedSensuClientURL() string {
	return fmt.Sprintf("%s.%s.svc", k8sutil.APIServiceName(s.clusterName), s.namespace)
}

func (s *SensuClient) CLI() *cli.SensuCli {
	s.ensureCredentials()
	return s.sensuCli
}

func (s *SensuClient) ensureCredentials() (err error) {
	var tokens *types.Tokens

	currentTokens := s.sensuCli.Config.Tokens()
	s.logger.Warnf("currentTokens during ensureCredentials: %s", spew.Sdump(currentTokens))
	if currentTokens == nil || currentTokens.Access == "" {
		s.logger.Warnf("About to attempt to create access token with url: %s", fmt.Sprintf("http://%s:8080", s.makeFullyQualifiedSensuClientURL()))
		if tokens, err = s.sensuCli.Client.CreateAccessToken(fmt.Sprintf("http://%s:8080", s.makeFullyQualifiedSensuClientURL()), "admin", "P@ssw0rd!"); err != nil {
			s.logger.Errorf("create token err: %+v", err)
			return
		}

		conf := basic.Config{
			Cluster: basic.Cluster{
				APIUrl:  s.sensuCli.Config.APIUrl(),
				Edition: "enterprise",
				Tokens:  tokens,
			},
			Profile: basic.Profile{
				Format:    "json",
				Namespace: s.sensuCli.Config.Namespace(),
			},
		}

		sensuCliClient := client.New(&conf)

		logger := logrus.WithFields(logrus.Fields{
			"component": "cli-client",
		})

		s.sensuCli = &cli.SensuCli{
			Client: sensuCliClient,
			Config: &conf,
			Logger: logger,
		}
	}
	return nil
}
