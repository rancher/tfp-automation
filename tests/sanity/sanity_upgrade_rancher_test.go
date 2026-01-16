package sanity

import (
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/infrastructure/ranchers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TfpSanityUpgradeRancherTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	session                    *session.Session
	cattleConfig               map[string]any
	rancherConfig              *rancher.Config
	terraformConfig            *config.TerraformConfig
	terratestConfig            *config.TerratestConfig
	standaloneConfig           *config.Standalone
	standaloneTerraformOptions *terraform.Options
	upgradeTerraformOptions    *terraform.Options
	terraformOptions           *terraform.Options
	serverNodeOne              string
}

func (s *TfpSanityUpgradeRancherTestSuite) TearDownSuite() {
	_, keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath, s.terratestConfig.PathToRepo, s.terraformConfig.Provider)
	cleanup.Cleanup(s.T(), s.standaloneTerraformOptions, keyPath)

	_, keyPath = rancher2.SetKeyPath(keypath.UpgradeKeyPath, s.terratestConfig.PathToRepo, s.terraformConfig.Provider)
	cleanup.Cleanup(s.T(), s.upgradeTerraformOptions, keyPath)
}

func (s *TfpSanityUpgradeRancherTestSuite) SetupSuite() {
	testSession := session.NewSession()
	s.session = testSession

	s.client, s.serverNodeOne, s.standaloneTerraformOptions, s.terraformOptions, s.cattleConfig = ranchers.SetupRancher(s.T(), s.session, keypath.SanityKeyPath)
	s.rancherConfig, s.terraformConfig, s.terratestConfig, s.standaloneConfig = config.LoadTFPConfigs(s.cattleConfig)
}

func (s *TfpSanityUpgradeRancherTestSuite) TestTfpUpgradeRancher() {
	var clusterIDs []string

	standardUserClient, newFile, rootBody, file, standardToken, testUser, testPassword := ranchers.SetupResources(s.T(), s.client, s.rancherConfig, s.terratestConfig, s.terraformOptions)

	s.rancherConfig, s.terraformConfig, s.terratestConfig, _ = config.LoadTFPConfigs(s.cattleConfig)
	allClusterIDs := s.provisionAndVerifyCluster("Sanity_Pre_Rancher_Upgrade_", clusterIDs, newFile, rootBody, file, standardUserClient, standardToken, testUser, testPassword)

	s.client, s.cattleConfig, s.terraformOptions, s.upgradeTerraformOptions = ranchers.UpgradeRancher(s.T(), s.client, s.serverNodeOne, s.session, s.cattleConfig)
	provisioning.VerifyClustersState(s.T(), s.client, allClusterIDs)

	ranchers.CleanupPreUpgradeClusters(s.T(), s.client, allClusterIDs, s.terraformConfig)

	standardUserClient, newFile, rootBody, file, standardToken, testUser, testPassword = ranchers.SetupResources(s.T(), s.client, s.rancherConfig, s.terratestConfig, s.terraformOptions)

	s.rancherConfig, s.terraformConfig, s.terratestConfig, _ = config.LoadTFPConfigs(s.cattleConfig)
	s.provisionAndVerifyCluster("Sanity_Post_Rancher_Upgrade_", nil, newFile, rootBody, file, standardUserClient, standardToken, testUser, testPassword)

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, s.terratestConfig.PathToRepo, "")
	cleanup.Cleanup(s.T(), s.terraformOptions, keyPath)

	if s.terratestConfig.LocalQaseReporting {
		results.ReportTest(s.terratestConfig)
	}
}

func (s *TfpSanityUpgradeRancherTestSuite) provisionAndVerifyCluster(name string, clusterIDs []string, newFile *hclwrite.File, rootBody *hclwrite.Body,
	file *os.File, standardUserClient *rancher.Client, standardToken, testUser, testPassword string) []string {
	customClusterNames := []string{}
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}
	rke2Module, rke2Windows2019, rke2Windows2022, k3sModule := provisioning.DownstreamClusterModules(s.terraformConfig)

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		module    string
	}{
		{"RKE2", nodeRolesDedicated, rke2Module},
		{"RKE2_Windows_2019", nil, rke2Windows2019},
		{"RKE2_Windows_2022", nil, rke2Windows2022},
		{"K3S", nodeRolesDedicated, k3sModule},
	}

	for _, tt := range tests {
		if strings.Contains(tt.name, "Windows") && (s.terraformConfig.Provider != aws.Aws) {
			s.T().Skip("Skipping Windows test on non-AWS provider")
		}

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{s.cattleConfig})
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, configMap[0])
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(s.T(), err)

		provisioning.GetK8sVersion(s.T(), standardUserClient, s.terratestConfig, s.terraformConfig, configMap)

		rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

		tt.name = name + tt.name

		s.Run((tt.name), func() {
			clusterIDs, customClusterNames = provisioning.Provision(s.T(), s.client, standardUserClient, rancher, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file, false, true, true, clusterIDs, customClusterNames)
			provisioning.VerifyClustersState(s.T(), s.client, clusterIDs)
			provisioning.VerifyServiceAccountTokenSecret(s.T(), s.client, clusterIDs)

			cluster, err := s.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraform.ResourcePrefix)
			require.NoError(s.T(), err)

			err = pods.VerifyClusterPods(s.client, cluster)
			require.NoError(s.T(), err)

			if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
				clusterIDs, customClusterNames = provisioning.Provision(s.T(), s.client, standardUserClient, rancher, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file, true, true, true, clusterIDs, customClusterNames)
				provisioning.VerifyClustersState(s.T(), s.client, clusterIDs)
				provisioning.VerifyServiceAccountTokenSecret(s.T(), s.client, clusterIDs)

				err = pods.VerifyClusterPods(s.client, cluster)
				require.NoError(s.T(), err)
			}
		})

		params := tfpQase.GetProvisioningSchemaParams(configMap[0])
		err = qase.UpdateSchemaParameters(tt.name, params)
		if err != nil {
			logrus.Warningf("Failed to upload schema parameters %s", err)
		}
	}

	return ranchers.UniqueStrings(clusterIDs)
}

func TestTfpSanityUpgradeRancherTestSuite(t *testing.T) {
	suite.Run(t, new(TfpSanityUpgradeRancherTestSuite))
}
