package nodescaling

import (
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
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
	terratestConfig  *config.TerratestConfig
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

	terratestConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, terratestConfig)

	s.terratestConfig = terratestConfig

	keyPath := rancher2.SetKeyPath()
	terraformOptions := framework.Setup(s.T(), s.terraformConfig, s.terratestConfig, keyPath)
	s.terraformOptions = terraformOptions
}

func (s *ScaleHostedTestSuite) TestTfpScaleHosted() {
	tests := []struct {
		name string
	}{
		{config.StandardClientName.String()},
	}

	for _, tt := range tests {
		tt.name = tt.name + " Module: " + s.terraformConfig.Module + " Kubernetes version: " + s.terratestConfig.KubernetesVersion

		testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

		s.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath()
			defer cleanup.Cleanup(s.T(), s.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(s.T(), s.client)
			require.NoError(s.T(), err)

			provisioning.Provision(s.T(), s.client, s.rancherConfig, s.terraformConfig, s.terratestConfig, testUser, testPassword, clusterName, poolName, s.terraformOptions, nil)
			provisioning.VerifyCluster(s.T(), adminClient, clusterName, s.terraformConfig, s.terratestConfig)

			s.terratestConfig.Nodepools = s.terratestConfig.ScalingInput.ScaledUpNodepools

			provisioning.Scale(s.T(), s.client, s.rancherConfig, s.terraformConfig, s.terratestConfig, testUser, testPassword, clusterName, poolName, s.terraformOptions)

			time.Sleep(4 * time.Minute)

			provisioning.VerifyCluster(s.T(), adminClient, clusterName, s.terraformConfig, s.terratestConfig)
			provisioning.VerifyNodeCount(s.T(), s.client, clusterName, s.terraformConfig, s.terratestConfig.ScalingInput.ScaledUpNodeCount)

			s.terratestConfig.Nodepools = s.terratestConfig.ScalingInput.ScaledDownNodepools

			provisioning.Scale(s.T(), s.client, s.rancherConfig, s.terraformConfig, s.terratestConfig, testUser, testPassword, clusterName, poolName, s.terraformOptions)

			time.Sleep(4 * time.Minute)

			provisioning.VerifyCluster(s.T(), adminClient, clusterName, s.terraformConfig, s.terratestConfig)
			provisioning.VerifyNodeCount(s.T(), s.client, clusterName, s.terraformConfig, s.terratestConfig.ScalingInput.ScaledDownNodeCount)
		})
	}

	if s.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpScaleHostedTestSuite(t *testing.T) {
	suite.Run(t, new(ScaleHostedTestSuite))
}
