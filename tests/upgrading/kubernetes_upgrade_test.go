package tests

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/framework"
	cleanup "github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type KubernetesUpgradeTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	terraformConfig  *config.TerraformConfig
	clusterConfig    *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (k *KubernetesUpgradeTestSuite) SetupSuite() {
	testSession := session.NewSession()
	k.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(k.T(), err)

	k.client = client

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, terraformConfig)

	k.terraformConfig = terraformConfig

	clusterConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, clusterConfig)

	k.clusterConfig = clusterConfig

	terraformOptions := framework.Setup(k.T())
	k.terraformOptions = terraformOptions

	provisioning.GetK8sVersion(k.T(), k.client, k.clusterConfig, k.terraformConfig, configs.SecondHighestVersion)
}

func (k *KubernetesUpgradeTestSuite) TestTfpKubernetesUpgrade() {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
	}{
		{"3 nodes - 1 role per node " + config.StandardClientName.String(), nodeRolesDedicated},
	}

	for _, tt := range tests {
		clusterConfig := *k.clusterConfig
		clusterConfig.Nodepools = tt.nodeRoles

		tt.name = tt.name + " Module: " + k.terraformConfig.Module + " Kubernetes version: " + k.clusterConfig.KubernetesVersion

		clusterName := namegen.AppendRandomString(configs.TFP)
		poolName := namegen.AppendRandomString(configs.TFP)

		k.Run((tt.name), func() {
			defer cleanup.ConfigCleanup(k.T(), k.terraformOptions)

			provisioning.Provision(k.T(), clusterName, poolName, &clusterConfig, k.terraformOptions)
			provisioning.VerifyCluster(k.T(), k.client, clusterName, k.terraformConfig, k.terraformOptions, &clusterConfig)

			provisioning.KubernetesUpgrade(k.T(), k.client, clusterName, poolName, k.terraformOptions, k.terraformConfig, &clusterConfig)
			provisioning.VerifyCluster(k.T(), k.client, clusterName, k.terraformConfig, k.terraformOptions, &clusterConfig)
			provisioning.VerifyUpgradedKubernetesVersion(k.T(), k.client, k.terraformConfig, clusterName,
				k.clusterConfig.UpgradedKubernetesVersion)
		})
	}
}

func (k *KubernetesUpgradeTestSuite) TestTfpKubernetesUpgradeDynamicInput() {
	tests := []struct {
		name string
	}{
		{config.StandardClientName.String()},
	}

	for _, tt := range tests {
		tt.name = tt.name + " Module: " + k.terraformConfig.Module + " Kubernetes version: " + k.clusterConfig.KubernetesVersion

		clusterName := namegen.AppendRandomString(configs.TFP)
		poolName := namegen.AppendRandomString(configs.TFP)

		k.Run((tt.name), func() {
			defer cleanup.ConfigCleanup(k.T(), k.terraformOptions)

			provisioning.Provision(k.T(), clusterName, poolName, k.clusterConfig, k.terraformOptions)
			provisioning.VerifyCluster(k.T(), k.client, clusterName, k.terraformConfig, k.terraformOptions, k.clusterConfig)

			provisioning.KubernetesUpgrade(k.T(), k.client, clusterName, poolName, k.terraformOptions, k.terraformConfig, k.clusterConfig)
			provisioning.VerifyCluster(k.T(), k.client, clusterName, k.terraformConfig, k.terraformOptions, k.clusterConfig)
			provisioning.VerifyUpgradedKubernetesVersion(k.T(), k.client, k.terraformConfig, clusterName,
				k.clusterConfig.UpgradedKubernetesVersion)
		})
	}
}

func TestTfpKubernetesUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(KubernetesUpgradeTestSuite))
}
