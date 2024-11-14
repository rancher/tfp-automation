package nodescaling

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/framework"
	cleanup "github.com/rancher/tfp-automation/framework/cleanup"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ScaleHostedTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	rancherConfig    *rancher.Config
	terraformConfig  *config.TerraformConfig
	clusterConfig    *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (s *ScaleHostedTestSuite) SetupSuite() {
	testSession := session.NewSession()
	s.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(s.T(), err)

	s.client = client

	rancherConfig := new(rancher.Config)
	ranchFrame.LoadConfig(configs.Rancher, rancherConfig)

	s.rancherConfig = rancherConfig

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, terraformConfig)

	s.terraformConfig = terraformConfig

	clusterConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, clusterConfig)

	s.clusterConfig = clusterConfig

	terraformOptions := framework.Setup(s.T(), s.rancherConfig, s.terraformConfig, s.clusterConfig)
	s.terraformOptions = terraformOptions
}

func (s *ScaleHostedTestSuite) TestTfpScaleHosted() {
	tests := []struct {
		name string
	}{
		{config.StandardClientName.String()},
	}

	for _, tt := range tests {
		tt.name = tt.name + " Module: " + s.terraformConfig.Module + " Kubernetes version: " + s.clusterConfig.KubernetesVersion

		testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

		s.Run((tt.name), func() {
			defer cleanup.ConfigCleanup(s.T(), s.terraformOptions)

			provisioning.Provision(s.T(), s.client, s.rancherConfig, s.terraformConfig, s.clusterConfig, testUser, testPassword, clusterName, poolName, s.terraformOptions)
			provisioning.VerifyCluster(s.T(), s.client, clusterName, s.terraformConfig, s.terraformOptions, s.clusterConfig)

			s.clusterConfig.Nodepools = s.clusterConfig.ScalingInput.ScaledUpNodepools

			provisioning.Scale(s.T(), s.rancherConfig, s.terraformConfig, s.clusterConfig, testUser, testPassword, clusterName, poolName, s.terraformOptions)
			provisioning.VerifyCluster(s.T(), s.client, clusterName, s.terraformConfig, s.terraformOptions, s.clusterConfig)
			provisioning.VerifyNodeCount(s.T(), s.client, clusterName, s.terraformConfig, s.clusterConfig.ScalingInput.ScaledUpNodeCount)

			s.clusterConfig.Nodepools = s.clusterConfig.ScalingInput.ScaledDownNodepools

			provisioning.Scale(s.T(), s.rancherConfig, s.terraformConfig, s.clusterConfig, testUser, testPassword, clusterName, poolName, s.terraformOptions)
			provisioning.VerifyCluster(s.T(), s.client, clusterName, s.terraformConfig, s.terraformOptions, s.clusterConfig)
			provisioning.VerifyNodeCount(s.T(), s.client, clusterName, s.terraformConfig, s.clusterConfig.ScalingInput.ScaledDownNodeCount)
		})
	}

	if s.clusterConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpScaleHostedTestSuite(t *testing.T) {
	t.Skip("This test has been deprecated.")
	suite.Run(t, new(ScaleHostedTestSuite))
}
