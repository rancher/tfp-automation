package sanity

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	clusterActions "github.com/rancher/tests/actions/clusters"
	provisioningActions "github.com/rancher/tests/actions/provisioning"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
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

type TfpSanityDualStackUpgradeRancherTestSuite struct {
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

func (s *TfpSanityDualStackUpgradeRancherTestSuite) TearDownSuite() {
	_, keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath, s.terratestConfig.PathToRepo, s.terraformConfig.Provider)
	cleanup.Cleanup(s.T(), s.standaloneTerraformOptions, keyPath)

	_, keyPath = rancher2.SetKeyPath(keypath.UpgradeKeyPath, s.terratestConfig.PathToRepo, s.terraformConfig.Provider)
	cleanup.Cleanup(s.T(), s.upgradeTerraformOptions, keyPath)
}

func (s *TfpSanityDualStackUpgradeRancherTestSuite) SetupSuite() {
	testSession := session.NewSession()
	s.session = testSession

	s.client, s.serverNodeOne, s.standaloneTerraformOptions, s.terraformOptions, s.cattleConfig = ranchers.SetupDualStackRancher(s.T(), s.session, keypath.DualStackKeyPath)
	s.rancherConfig, s.terraformConfig, s.terratestConfig, s.standaloneConfig = config.LoadTFPConfigs(s.cattleConfig)
}

func (s *TfpSanityDualStackUpgradeRancherTestSuite) TestTfpUpgradeDualStackRancher() {
	standardUserClient, standardToken, testUser, testPassword := ranchers.SetupResources(s.T(), s.client, s.rancherConfig, s.terratestConfig, s.terraformOptions)

	s.rancherConfig, s.terraformConfig, s.terratestConfig, _ = config.LoadTFPConfigs(s.cattleConfig)
	nestedRancherModuleDir := s.provisionAndVerifyCluster("Sanity_DualStack_Pre_Rancher_Upgrade", standardUserClient, standardToken, testUser, testPassword)

	s.client, s.cattleConfig, s.terraformOptions, s.upgradeTerraformOptions = ranchers.UpgradeDualStackRancher(s.T(), s.client, s.serverNodeOne, s.session, s.cattleConfig)

	ranchers.CleanupDownstreamClusters(s.T(), s.client, s.terraformConfig)
	os.RemoveAll(nestedRancherModuleDir)

	standardUserClient, standardToken, testUser, testPassword = ranchers.SetupResources(s.T(), s.client, s.rancherConfig, s.terratestConfig, s.terraformOptions)

	s.rancherConfig, s.terraformConfig, s.terratestConfig, _ = config.LoadTFPConfigs(s.cattleConfig)
	nestedRancherModuleDir = s.provisionAndVerifyCluster("Sanity_DualStack_Post_Rancher_Upgrade", standardUserClient, standardToken, testUser, testPassword)

	ranchers.CleanupDownstreamClusters(s.T(), s.client, s.terraformConfig)
	os.RemoveAll(nestedRancherModuleDir)

	if s.terratestConfig.LocalQaseReporting {
		results.ReportTest(s.terratestConfig)
	}
}

func (s *TfpSanityDualStackUpgradeRancherTestSuite) provisionAndVerifyCluster(name string, standardUserClient *rancher.Client, standardToken,
	testUser, testPassword string) string {
	var clusterIDs []string
	var nestedRancherModuleDir string

	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}
	rke2Module, _, _, k3sModule := provisioning.DownstreamClusterModules(s.terraformConfig)

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		module    string
	}{
		{name + "_RKE2", nodeRolesDedicated, rke2Module},
		{name + "_K3S", nodeRolesDedicated, k3sModule},
	}

	s.T().Run(name, func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(s.terraformConfig, s.terratestConfig, s.terraformOptions, tt.name, configs.NestedRancherModuleDir)
				require.NoError(t, err)

				newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
				defer file.Close()

				cattleConfig, err := provisioning.UniquifyTerraform(s.cattleConfig)
				require.NoError(t, err)

				_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, cattleConfig)
				require.NoError(t, err)

				_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, cattleConfig)
				require.NoError(t, err)

				_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, cattleConfig)
				require.NoError(t, err)

				provisioning.GetK8sVersion(standardUserClient, cattleConfig)

				rancher, terraform, terratest, _ := config.LoadTFPConfigs(cattleConfig)

				clusters, _ := provisioning.Provision(t, s.client, standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, newFile, rootBody, file, false, true, true, clusterIDs, nil, nestedRancherModuleDir)
				err = provisioningActions.VerifyClusterReady(s.client, clusters[0])
				require.NoError(t, err)

				err = clusterActions.VerifyServiceAccountTokenSecret(s.client, clusters[0].Name)
				require.NoError(t, err)
				cluster, err := s.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraform.ResourcePrefix)
				require.NoError(t, err)

				err = pods.VerifyClusterPods(s.client, cluster)
				require.NoError(t, err)

				params := tfpQase.GetProvisioningSchemaParams(cattleConfig)
				err = qase.UpdateSchemaParameters(tt.name, params)
				if err != nil {
					logrus.Warningf("Failed to upload schema parameters %s", err)
				}
			})
		}
	})

	return nestedRancherModuleDir
}

func TestTfpSanityDualStackUpgradeRancherTestSuite(t *testing.T) {
	suite.Run(t, new(TfpSanityDualStackUpgradeRancherTestSuite))
}
