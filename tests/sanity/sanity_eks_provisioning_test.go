package sanity

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/infrastructure/ranchers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TfpSanityEKSProvisioningTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	standardUserClient         *rancher.Client
	session                    *session.Session
	cattleConfig               map[string]any
	rancherConfig              *rancher.Config
	terraformConfig            *config.TerraformConfig
	terratestConfig            *config.TerratestConfig
	standaloneConfig           *config.Standalone
	standaloneTerraformOptions *terraform.Options
	terraformOptions           *terraform.Options
}

func (s *TfpSanityEKSProvisioningTestSuite) SetupSuite() {
	testSession := session.NewSession()
	s.session = testSession

	s.client, _, s.standaloneTerraformOptions, s.terraformOptions, s.cattleConfig = ranchers.SetupRancher(s.T(), s.session, keypath.HostedKeyPath)
	s.rancherConfig, s.terraformConfig, s.terratestConfig, s.standaloneConfig = config.LoadTFPConfigs(s.cattleConfig)
}

func (s *TfpSanityEKSProvisioningTestSuite) TestTfpProvisioningEKSSanity() {
	var err error
	var testUser, testPassword string
	var clusterIDs []string

	customClusterNames := []string{}

	s.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(s.client)
	require.NoError(s.T(), err)

	standardUserToken, err := ranchers.CreateStandardUserToken(s.T(), s.terraformOptions, s.rancherConfig, testUser, testPassword)
	require.NoError(s.T(), err)

	standardToken := standardUserToken.Token

	eksNodePools := []config.Nodepool{{DiskSize: 100, InstanceType: s.terraformConfig.AWSConfig.AWSInstanceType, DesiredSize: 3, MaxSize: 3, MinSize: 3}}

	tests := []struct {
		name              string
		module            string
		nodePools         []config.Nodepool
		kubernetesVersion string
	}{
		{"Sanity_EKS", modules.EKS, eksNodePools, s.terratestConfig.EKSKubernetesVersion},
	}

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(s.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{s.cattleConfig})
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, configMap[0])
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodePools, configMap[0])
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "kubernetesVersion"}, tt.kubernetesVersion, configMap[0])
		require.NoError(s.T(), err)

		provisioning.GetK8sVersion(s.T(), s.standardUserClient, s.terratestConfig, s.terraformConfig, configMap)

		rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

		s.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, s.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(s.T(), s.terraformOptions, keyPath)

			clusterIDs, _ := provisioning.Provision(s.T(), s.client, s.standardUserClient, rancher, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file, false, false, true, clusterIDs, customClusterNames)
			provisioning.VerifyClustersState(s.T(), s.client, clusterIDs)
			provisioning.VerifyServiceAccountTokenSecret(s.T(), s.client, clusterIDs)
			provisioning.VerifyV3ClustersPods(s.T(), s.client, clusterIDs)
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

func TestTfpSanityEKSProvisioningTestSuite(t *testing.T) {
	suite.Run(t, new(TfpSanityEKSProvisioningTestSuite))
}
