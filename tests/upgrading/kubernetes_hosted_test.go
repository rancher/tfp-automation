package tests

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
	cleanup "github.com/rancher/tfp-automation/framework/cleanup"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type KubernetesUpgradeHostedTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	rancherConfig    *rancher.Config
	terraformConfig  *config.TerraformConfig
	clusterConfig    *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (k *KubernetesUpgradeHostedTestSuite) SetupSuite() {
	testSession := session.NewSession()
	k.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(k.T(), err)

	k.client = client

	rancherConfig := new(rancher.Config)
	ranchFrame.LoadConfig(configs.Rancher, rancherConfig)

	k.rancherConfig = rancherConfig

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, terraformConfig)

	k.terraformConfig = terraformConfig

	clusterConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, clusterConfig)

	k.clusterConfig = clusterConfig

	terraformOptions := framework.Setup(k.T(), k.rancherConfig, k.terraformConfig, k.clusterConfig)
	k.terraformOptions = terraformOptions
}

func (k *KubernetesUpgradeHostedTestSuite) TestTfpKubernetesUpgradeHosted() {
	tests := []struct {
		name string
	}{
		{config.StandardClientName.String()},
	}

	for _, tt := range tests {
		tt.name = tt.name + " Module: " + k.terraformConfig.Module + " Kubernetes version: " + k.clusterConfig.KubernetesVersion

		testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

		k.Run((tt.name), func() {
			defer cleanup.ConfigCleanup(k.T(), k.terraformOptions)

			adminClient, err := provisioning.FetchAdminClient(k.T(), k.client)
			require.NoError(k.T(), err)

			provisioning.Provision(k.T(), k.client, k.rancherConfig, k.terraformConfig, k.clusterConfig, testUser, testPassword, clusterName, poolName, k.terraformOptions, nil)
			provisioning.VerifyCluster(k.T(), adminClient, clusterName, k.terraformConfig, k.terraformOptions, k.clusterConfig)

			provisioning.KubernetesUpgrade(k.T(), k.client, k.rancherConfig, k.terraformConfig, k.clusterConfig, testUser, testPassword, clusterName, poolName, k.terraformOptions)

			time.Sleep(4 * time.Minute)

			provisioning.VerifyCluster(k.T(), adminClient, clusterName, k.terraformConfig, k.terraformOptions, k.clusterConfig)
			provisioning.VerifyUpgradedKubernetesVersion(k.T(), k.client, k.terraformConfig, clusterName, k.clusterConfig.UpgradedKubernetesVersion)
		})
	}

	if k.clusterConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpKubernetesUpgradeHostedTestSuite(t *testing.T) {
	suite.Run(t, new(KubernetesUpgradeHostedTestSuite))
}
