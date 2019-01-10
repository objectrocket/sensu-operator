package client

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/client/config/basic"
	"github.com/sirupsen/logrus"
)

func TestNew(t *testing.T) {
	conf := basic.Config{
		Cluster: basic.Cluster{
			APIUrl:  "http://testCluster.testnamespace.svc:8080",
			Edition: "enterprise",
		},
		Profile: basic.Profile{
			Format:    "json",
			Namespace: "testnamespace",
		},
	}
	sensuCliClient := client.New(&conf)
	logger := logrus.WithFields(logrus.Fields{
		"component": "cli-client",
	})

	type args struct {
		clusterName string
		namespace   string
	}
	tests := []struct {
		name string
		args args
		want *SensuClient
	}{
		{
			"valid",
			args{
				"testCluster",
				"testnamespace",
			},
			&SensuClient{
				logger:      logrus.WithField("pkg", "sensu_client").WithField("cluster-name", "testCluster"),
				clusterName: "testCluster",
				namespace:   "testnamespace",
				sensuCli: &cli.SensuCli{
					Client: sensuCliClient,
					Config: &conf,
					Logger: logger,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.clusterName, tt.args.namespace, tt.args.namespace); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", spew.Sdump(got), spew.Sdump(tt.want))
			}
		})
	}
}

func TestSensuClient_makeFullyQualifiedSensuClientURL(t *testing.T) {
	conf := basic.Config{
		Cluster: basic.Cluster{
			APIUrl:  "http://testCluster.testnamespace.svc:8080",
			Edition: "enterprise",
		},
		Profile: basic.Profile{
			Format:    "json",
			Namespace: "testnamespace",
		},
	}
	sensuCliClient := client.New(&conf)
	logger := logrus.WithFields(logrus.Fields{
		"component": "cli-client",
	})

	type fields struct {
		logger      *logrus.Entry
		clusterName string
		namespace   string
		sensuCli    *cli.SensuCli
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"test",
			fields{
				logrus.WithField("pkg", "sensu_client").WithField("cluster-name", "testCluster"),
				"testCluster",
				"testnamespace",
				&cli.SensuCli{
					Client: sensuCliClient,
					Config: &conf,
					Logger: logger,
				},
			},
			"testCluster-api.testnamespace.svc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SensuClient{
				logger:      tt.fields.logger,
				clusterName: tt.fields.clusterName,
				namespace:   tt.fields.namespace,
				sensuCli:    tt.fields.sensuCli,
			}
			if got := s.makeFullyQualifiedSensuClientURL(); got != tt.want {
				t.Errorf("SensuClient.makeFullyQualifiedSensuClientURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
