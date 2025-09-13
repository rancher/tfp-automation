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
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/infrastructure"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ScaleTestSuite struct {
	suite.Suite
	client             *rancher.Client
	standardUserClient *rancher.Client
	session            *session.Session
	cattleConfig       map[string]any
	rancherConfig      *rancher.Config
	terraformConfig    *config.TerraformConfig
	terratestConfig    *config.TerratestConfig
	terraformOptions   *terraform.Options
}

func (s *ScaleTestSuite) SetupSuite() {
	testSession := session.NewSession()
	s.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(s.T(), err)

	s.client = client

	s.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	s.rancherConfig, s.terraformConfig, s.terratestConfig, _ = config.LoadTFPConfigs(s.cattleConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, s.terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(s.T(), s.terraformConfig, s.terratestConfig, keyPath)
	s.terraformOptions = terraformOptions
}

func (s *ScaleTestSuite) TestTfpScale() {
	var err error
	var testUser, testPassword string

	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}
	scaleUpRolesDedicated := []config.Nodepool{config.ScaleUpEtcdNodePool, config.ScaleUpControlPlaneNodePool, config.ScaleUpWorkerNodePool}
	scaleDownRolesDedicated := []config.Nodepool{config.ScaleUpEtcdNodePool, config.ScaleUpControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name               string
		nodeRoles          []config.Nodepool
		scaleUpNodeRoles   []config.Nodepool
		scaleDownNodeRoles []config.Nodepool
	}{
		{"Scaling_8_nodes_to_13_nodes_11_nodes", nodeRolesDedicated, scaleUpRolesDedicated, scaleDownRolesDedicated},
	}

	s.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(s.client)
	require.NoError(s.T(), err)

	standardUserToken, err := infrastructure.CreateStandardUserToken(s.T(), s.terraformOptions, s.rancherConfig, testUser, testPassword)
	require.NoError(s.T(), err)

	standardToken := standardUserToken.Token

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(s.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{s.cattleConfig})
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, configMap[0])
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		require.NoError(s.T(), err)

		provisioning.GetK8sVersion(s.T(), s.client, s.terratestConfig, s.terraformConfig, configs.DefaultK8sVersion, configMap)

		rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

		var scaledUpCount, scaledDownCount int64

		for _, scaleUpNodepool := range tt.scaleUpNodeRoles {
			scaledUpCount += scaleUpNodepool.Quantity
		}

		for _, scaleDownNodepool := range tt.scaleDownNodeRoles {
			scaledDownCount += scaleDownNodepool.Quantity
		}

		s.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, s.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(s.T(), s.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(s.T(), s.client)
			require.NoError(s.T(), err)

			clusterIDs, _ := provisioning.Provision(s.T(), s.client, s.standardUserClient, rancher, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file, false, false, false, nil)
			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)

			_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.scaleUpNodeRoles, configMap[0])
			require.NoError(s.T(), err)

			provisioning.Scale(s.T(), s.standardUserClient, rancher, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file)

			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)
			provisioning.VerifyNodeCount(s.T(), adminClient, terraform.ResourcePrefix, terraform, scaledUpCount)

			_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.scaleDownNodeRoles, configMap[0])
			require.NoError(s.T(), err)

			provisioning.Scale(s.T(), s.standardUserClient, rancher, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file)

			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)
			provisioning.VerifyNodeCount(s.T(), adminClient, terraform.ResourcePrefix, terraform, scaledDownCount)
		})

		params := tfpQase.GetProvisioningSchemaParams(configMap[0])
		err = qase.UpdateSchemaParameters(tt.name, params)
		if err != nil {
			logrus.Warningf("Failed to upload schema parameters %s", err)
		}
	}

	if s.terratestConfig.LocalQaseReporting {
		results.ReportTest(s.terratestConfig)
	}
}

func (s *ScaleTestSuite) TestTfpScaleDynamicInput() {
	var err error
	var testUser, testPassword string

	tests := []struct {
		name string
	}{
		{config.StandardClientName.String()},
	}

	s.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(s.client)
	require.NoError(s.T(), err)

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(s.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{s.cattleConfig})
		require.NoError(s.T(), err)

		provisioning.GetK8sVersion(s.T(), s.client, s.terratestConfig, s.terraformConfig, configs.DefaultK8sVersion, configMap)

		rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

		s.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, s.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(s.T(), s.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(s.T(), s.client)
			require.NoError(s.T(), err)

			clusterIDs, _ := provisioning.Provision(s.T(), s.client, s.standardUserClient, rancher, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file, false, false, false, nil)
			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)

			operations.ReplaceValue([]string{"terratest", "nodepools"}, terratest.ScalingInput.ScaledUpNodepools, configMap[0])

			provisioning.Scale(s.T(), s.standardUserClient, rancher, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file)

			time.Sleep(2 * time.Minute)

			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)
			provisioning.VerifyNodeCount(s.T(), adminClient, terraform.ResourcePrefix, terraform, terratest.ScalingInput.ScaledUpNodeCount)

			operations.ReplaceValue([]string{"terratest", "nodepools"}, terratest.ScalingInput.ScaledDownNodepools, configMap[0])

			provisioning.Scale(s.T(), s.client, rancher, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file)

			time.Sleep(2 * time.Minute)

			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)
			provisioning.VerifyNodeCount(s.T(), adminClient, terraform.ResourcePrefix, s.terraformConfig, terratest.ScalingInput.ScaledDownNodeCount)
		})

		params := tfpQase.GetProvisioningSchemaParams(configMap[0])
		err = qase.UpdateSchemaParameters(tt.name, params)
		if err != nil {
			logrus.Warningf("Failed to upload schema parameters %s", err)
		}
	}

	if s.terratestConfig.LocalQaseReporting {
		results.ReportTest(s.terratestConfig)
	}
}

func TestTfpScaleTestSuite(t *testing.T) {
	suite.Run(t, new(ScaleTestSuite))
}
