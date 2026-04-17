package sanity

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/session"
	clusterActions "github.com/rancher/tests/actions/clusters"
	provisioningActions "github.com/rancher/tests/actions/provisioning"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	nested "github.com/rancher/tfp-automation/tests/extensions/nestedModules"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/infrastructure/ranchers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TfpSanityAKSProvisioningTestSuite struct {
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

func (s *TfpSanityAKSProvisioningTestSuite) SetupSuite() {
	testSession := session.NewSession()
	s.session = testSession

	s.client, _, s.standaloneTerraformOptions, s.terraformOptions, s.cattleConfig = ranchers.SetupRancher(s.T(), s.session, keypath.HostedKeyPath)
	s.rancherConfig, s.terraformConfig, s.terratestConfig, s.standaloneConfig = config.LoadTFPConfigs(s.cattleConfig)
}

func (s *TfpSanityAKSProvisioningTestSuite) TestTfpProvisioningAKSSanity() {
	var err error
	var testUser, testPassword string
	var clusterIDs []string

	customClusterName := ""

	s.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(s.client)
	require.NoError(s.T(), err)

	standardUserToken, err := ranchers.CreateStandardUserToken(s.T(), s.terraformOptions, s.rancherConfig, testUser, testPassword)
	require.NoError(s.T(), err)

	standardToken := standardUserToken.Token

	aksNodePools := []config.Nodepool{{Quantity: 3}}

	tests := []struct {
		name              string
		module            string
		nodePools         []config.Nodepool
		kubernetesVersion string
	}{
		{"Sanity_AKS", modules.HostedAzureAKS, aksNodePools, s.terratestConfig.AKSKubernetesVersion},
	}

	for _, tt := range tests {
		s.T().Skip("Skipping test - resource quota needs to be increased so GHA workflow properly runs")

		s.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(s.cattleConfig)
			rancher.AdminToken = standardToken
			terraform.Module = tt.module
			terratest.Nodepools = tt.nodePools
			terratest.KubernetesVersion = tt.kubernetesVersion

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(s.terraformConfig, s.terratestConfig, s.terraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()
			terratest, err = provisioning.GetK8sVersion(s.standardUserClient, terraform, terratest)
			require.NoError(s.T(), err)

			terraform = provisioning.UniquifyTerraform(terraform)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, s.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(s.T(), perTestTerraformOptions, keyPath)

			clusters, _ := provisioning.Provision(s.T(), s.client, s.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, newFile, rootBody, file, false, false, true, clusterIDs, customClusterName, nestedRancherModuleDir)
			err = provisioningActions.VerifyClusterReady(s.client, clusters[0])
			require.NoError(s.T(), err)

			err = clusterActions.VerifyServiceAccountTokenSecret(s.client, clusters[0].Name)
			require.NoError(s.T(), err)

			err = pods.VerifyClusterPods(s.client, clusters[0])
			require.NoError(s.T(), err)

			params := tfpQase.GetProvisioningSchemaParams(s.terraformConfig, s.terratestConfig)
			err = qase.UpdateSchemaParameters(tt.name, params)
			if err != nil {
				logrus.Warningf("Failed to upload schema parameters %s", err)
			}
		})
	}

	if s.terratestConfig.LocalQaseReporting {
		results.ReportTest(s.terratestConfig)
	}
}

func TestTfpSanityAKSProvisioningTestSuite(t *testing.T) {
	suite.Run(t, new(TfpSanityAKSProvisioningTestSuite))
}
