package nodescaling

import (
	"os"
	"testing"
	"time"

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

type ScaleTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	cattleConfig     map[string]any
	rancherConfig    *rancher.Config
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (s *ScaleTestSuite) SetupSuite() {
	testSession := session.NewSession()
	s.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(s.T(), err)

	s.client = client

	s.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	configMap, err := provisioning.UniquifyTerraform([]map[string]any{s.cattleConfig})
	require.NoError(s.T(), err)

	s.cattleConfig = configMap[0]
	s.rancherConfig, s.terraformConfig, s.terratestConfig = config.LoadTFPConfigs(s.cattleConfig)

	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
	terraformOptions := framework.Setup(s.T(), s.terraformConfig, s.terratestConfig, keyPath)
	s.terraformOptions = terraformOptions

	provisioning.GetK8sVersion(s.T(), s.client, s.terratestConfig, s.terraformConfig, configs.DefaultK8sVersion, configMap)
}

func (s *ScaleTestSuite) TestTfpScale() {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}
	scaleUpRolesDedicated := []config.Nodepool{config.ScaleUpEtcdNodePool, config.ScaleUpControlPlaneNodePool, config.ScaleUpWorkerNodePool}
	scaleDownRolesDedicated := []config.Nodepool{config.ScaleUpEtcdNodePool, config.ScaleUpControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name               string
		nodeRoles          []config.Nodepool
		scaleUpNodeRoles   []config.Nodepool
		scaleDownNodeRoles []config.Nodepool
	}{
		{"Scaling 3 nodes dedicated roles -> 8 nodes -> 6 nodes " + config.StandardClientName.String(), nodeRolesDedicated, scaleUpRolesDedicated, scaleDownRolesDedicated},
	}

	for _, tt := range tests {
		configMap := []map[string]any{s.cattleConfig}

		operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])

		provisioning.GetK8sVersion(s.T(), s.client, s.terratestConfig, s.terraformConfig, configs.DefaultK8sVersion, configMap)

		_, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		var scaledUpCount, scaledDownCount int64

		for _, scaleUpNodepool := range tt.scaleUpNodeRoles {
			scaledUpCount += scaleUpNodepool.Quantity
		}

		for _, scaleDownNodepool := range tt.scaleDownNodeRoles {
			scaledDownCount += scaleDownNodepool.Quantity
		}

		tt.name = tt.name + " Module: " + s.terraformConfig.Module + " Kubernetes version: " + terratest.KubernetesVersion

		testUser, testPassword := configs.CreateTestCredentials()

		s.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
			defer cleanup.Cleanup(s.T(), s.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(s.T(), s.client)
			require.NoError(s.T(), err)

			clusterIDs := provisioning.Provision(s.T(), s.client, s.rancherConfig, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, false)
			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)

			operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.scaleUpNodeRoles, configMap[0])

			provisioning.Scale(s.T(), s.client, s.rancherConfig, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap)

			time.Sleep(2 * time.Minute)

			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)
			provisioning.VerifyNodeCount(s.T(), s.client, terraform.ResourcePrefix, terraform, scaledUpCount)

			operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.scaleDownNodeRoles, configMap[0])

			provisioning.Scale(s.T(), s.client, s.rancherConfig, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap)

			time.Sleep(2 * time.Minute)

			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)
			provisioning.VerifyNodeCount(s.T(), s.client, terraform.ResourcePrefix, s.terraformConfig, scaledDownCount)
		})
	}

	if s.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func (s *ScaleTestSuite) TestTfpScaleDynamicInput() {
	tests := []struct {
		name string
	}{
		{config.StandardClientName.String()},
	}

	for _, tt := range tests {
		tt.name = tt.name + " Module: " + s.terraformConfig.Module + " Kubernetes version: " + s.terratestConfig.KubernetesVersion

		testUser, testPassword := configs.CreateTestCredentials()

		s.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
			defer cleanup.Cleanup(s.T(), s.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(s.T(), s.client)
			require.NoError(s.T(), err)

			configMap := []map[string]any{s.cattleConfig}

			clusterIDs := provisioning.Provision(s.T(), s.client, s.rancherConfig, s.terraformConfig, s.terratestConfig, testUser, testPassword, s.terraformOptions, configMap, false)
			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)

			operations.ReplaceValue([]string{"terratest", "nodepools"}, s.terratestConfig.ScalingInput.ScaledUpNodepools, configMap[0])

			provisioning.Scale(s.T(), s.client, s.rancherConfig, s.terraformConfig, s.terratestConfig, testUser, testPassword, s.terraformOptions, configMap)

			time.Sleep(2 * time.Minute)

			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)
			provisioning.VerifyNodeCount(s.T(), adminClient, s.terraformConfig.ResourcePrefix, s.terraformConfig, s.terratestConfig.ScalingInput.ScaledUpNodeCount)

			operations.ReplaceValue([]string{"terratest", "nodepools"}, s.terratestConfig.ScalingInput.ScaledDownNodepools, configMap[0])

			provisioning.Scale(s.T(), s.client, s.rancherConfig, s.terraformConfig, s.terratestConfig, testUser, testPassword, s.terraformOptions, configMap)

			time.Sleep(2 * time.Minute)

			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)
			provisioning.VerifyNodeCount(s.T(), adminClient, s.terraformConfig.ResourcePrefix, s.terraformConfig, s.terratestConfig.ScalingInput.ScaledDownNodeCount)
		})
	}

	if s.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpScaleTestSuite(t *testing.T) {
	suite.Run(t, new(ScaleTestSuite))
}
