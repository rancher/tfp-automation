package upgrading

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type KubernetesUpgradeTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	cattleConfig     map[string]any
	rancherConfig    *rancher.Config
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (k *KubernetesUpgradeTestSuite) SetupSuite() {
	testSession := session.NewSession()
	k.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(k.T(), err)

	k.client = client

	k.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	configMap, err := provisioning.UniquifyTerraform([]map[string]any{k.cattleConfig})
	require.NoError(k.T(), err)

	k.cattleConfig = configMap[0]
	k.rancherConfig, k.terraformConfig, k.terratestConfig = config.LoadTFPConfigs(k.cattleConfig)

	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, nil)
	terraformOptions := framework.Setup(k.T(), k.terraformConfig, k.terratestConfig, keyPath)
	k.terraformOptions = terraformOptions
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
		configMap := []map[string]any{k.cattleConfig}

		operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])

		provisioning.GetK8sVersion(k.T(), k.client, k.terratestConfig, k.terraformConfig, configs.DefaultK8sVersion, configMap)

		_, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		tt.name = tt.name + " Module: " + k.terraformConfig.Module + " Kubernetes version: " + terratest.KubernetesVersion

		testUser, testPassword := configs.CreateTestCredentials()

		k.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, nil)
			defer cleanup.Cleanup(k.T(), k.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(k.T(), k.client)
			require.NoError(k.T(), err)

			configMap := []map[string]any{k.cattleConfig}

			clusterIDs := provisioning.Provision(k.T(), k.client, k.rancherConfig, terraform, terratest, testUser, testPassword, k.terraformOptions, configMap, false)
			provisioning.VerifyClustersState(k.T(), adminClient, clusterIDs)

			provisioning.KubernetesUpgrade(k.T(), k.client, k.rancherConfig, terraform, terratest, testUser, testPassword, k.terraformOptions, configMap)
			provisioning.VerifyClustersState(k.T(), adminClient, clusterIDs)
			provisioning.VerifyKubernetesVersion(k.T(), k.client, clusterIDs[0], terratest.KubernetesVersion, k.terraformConfig.Module)
		})
	}

	if k.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func (k *KubernetesUpgradeTestSuite) TestTfpKubernetesUpgradeDynamicInput() {
	tests := []struct {
		name string
	}{
		{config.StandardClientName.String()},
	}

	for _, tt := range tests {
		configMap := []map[string]any{k.cattleConfig}
		provisioning.GetK8sVersion(k.T(), k.client, k.terratestConfig, k.terraformConfig, configs.DefaultK8sVersion, configMap)

		tt.name = tt.name + " Module: " + k.terraformConfig.Module + " Kubernetes version: " + k.terratestConfig.KubernetesVersion

		testUser, testPassword := configs.CreateTestCredentials()

		k.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, nil)
			defer cleanup.Cleanup(k.T(), k.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(k.T(), k.client)
			require.NoError(k.T(), err)

			configMap := []map[string]any{k.cattleConfig}

			clusterIDs := provisioning.Provision(k.T(), k.client, k.rancherConfig, k.terraformConfig, k.terratestConfig, testUser, testPassword, k.terraformOptions, configMap, false)
			provisioning.VerifyClustersState(k.T(), adminClient, clusterIDs)

			provisioning.KubernetesUpgrade(k.T(), k.client, k.rancherConfig, k.terraformConfig, k.terratestConfig, testUser, testPassword, k.terraformOptions, configMap)
			provisioning.VerifyClustersState(k.T(), adminClient, clusterIDs)
			provisioning.VerifyKubernetesVersion(k.T(), k.client, clusterIDs[0], k.terratestConfig.KubernetesVersion, k.terraformConfig.Module)
		})
	}

	if k.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpKubernetesUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(KubernetesUpgradeTestSuite))
}
