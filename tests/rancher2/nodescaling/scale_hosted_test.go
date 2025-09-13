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
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ScaleHostedTestSuite struct {
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

func (s *ScaleHostedTestSuite) SetupSuite() {
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

func (s *ScaleHostedTestSuite) TestTfpScaleHosted() {
	var err error
	var testUser, testPassword string

	tests := []struct {
		name string
	}{
		{"Scaling_Hosted_Cluster"},
	}

	s.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(s.client)
	require.NoError(s.T(), err)

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(s.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{s.cattleConfig})
		require.NoError(s.T(), err)

		s.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, s.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(s.T(), s.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(s.T(), s.client)
			require.NoError(s.T(), err)

			clusterIDs, _ := provisioning.Provision(s.T(), s.client, s.standardUserClient, s.rancherConfig, s.terraformConfig, s.terratestConfig, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file, false, false, false, nil)
			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)

			_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, s.terratestConfig.ScalingInput.ScaledUpNodepools, configMap[0])
			require.NoError(s.T(), err)

			provisioning.Scale(s.T(), s.standardUserClient, s.rancherConfig, s.terraformConfig, s.terratestConfig, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file)

			time.Sleep(4 * time.Minute)

			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)
			provisioning.VerifyNodeCount(s.T(), adminClient, s.terraformConfig.ResourcePrefix, s.terraformConfig, s.terratestConfig.ScalingInput.ScaledUpNodeCount)

			_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, s.terratestConfig.ScalingInput.ScaledDownNodepools, configMap[0])
			require.NoError(s.T(), err)

			provisioning.Scale(s.T(), s.standardUserClient, s.rancherConfig, s.terraformConfig, s.terratestConfig, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file)

			time.Sleep(4 * time.Minute)

			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)
			provisioning.VerifyNodeCount(s.T(), adminClient, s.terraformConfig.ResourcePrefix, s.terraformConfig, s.terratestConfig.ScalingInput.ScaledDownNodeCount)
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

func TestTfpScaleHostedTestSuite(t *testing.T) {
	suite.Run(t, new(ScaleHostedTestSuite))
}
