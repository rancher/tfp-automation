package sanity

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
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

type TfpSanityRKE1ProvisioningTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	session                    *session.Session
	cattleConfig               map[string]any
	rancherConfig              *rancher.Config
	terraformConfig            *config.TerraformConfig
	terratestConfig            *config.TerratestConfig
	standaloneConfig           *config.Standalone
	standaloneTerraformOptions *terraform.Options
	terraformOptions           *terraform.Options
}

func (s *TfpSanityRKE1ProvisioningTestSuite) TearDownSuite() {
	_, keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath, s.terratestConfig.PathToRepo, s.terraformConfig.Provider)
	cleanup.Cleanup(s.T(), s.standaloneTerraformOptions, keyPath)
}

func (s *TfpSanityRKE1ProvisioningTestSuite) SetupSuite() {
	testSession := session.NewSession()
	s.session = testSession

	s.client, _, s.standaloneTerraformOptions, s.terraformOptions, s.cattleConfig = infrastructure.SetupRancher(s.T(), s.session, keypath.SanityKeyPath)
	s.rancherConfig, s.terraformConfig, s.terratestConfig, s.standaloneConfig = config.LoadTFPConfigs(s.cattleConfig)

	provisioning.VerifyRancherVersion(s.T(), s.rancherConfig.Host, s.standaloneConfig.RancherTagVersion)
}

func (s *TfpSanityRKE1ProvisioningTestSuite) TestTfpRKE1ProvisioningSanity() {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		module    string
	}{
		{"Sanity_RKE1", nodeRolesDedicated, modules.EC2RKE1},
	}

	customClusterNames := []string{}
	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(s.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{s.cattleConfig})
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, s.client.RancherConfig.AdminToken, configMap[0])
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(s.T(), err)

		provisioning.GetK8sVersion(s.T(), s.client, s.terratestConfig, s.terraformConfig, configs.DefaultK8sVersion, configMap)

		rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion

		s.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, s.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(s.T(), s.terraformOptions, keyPath)

			clusterIDs, _ := provisioning.Provision(s.T(), s.client, rancher, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file, false, false, true, customClusterNames)
			provisioning.VerifyClustersState(s.T(), s.client, clusterIDs)
		})

		params := tfpQase.GetProvisioningSchemaParams(s.client, configMap[0])
		err = qase.UpdateSchemaParameters(tt.name, params)
		if err != nil {
			logrus.Warningf("Failed to upload schema parameters %s", err)
		}
	}

	if s.terratestConfig.LocalQaseReporting {
		results.ReportTest(s.terratestConfig)
	}
}

func TestTfpSanityRKE1ProvisioningTestSuite(t *testing.T) {
	suite.Run(t, new(TfpSanityRKE1ProvisioningTestSuite))
}
