package sanity

import (
	"os"
	"strings"
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
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
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

type TfpSanityProvisioningTestSuite struct {
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

func (s *TfpSanityProvisioningTestSuite) TearDownSuite() {
	_, keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath, s.terratestConfig.PathToRepo, s.terraformConfig.Provider)
	cleanup.Cleanup(s.T(), s.standaloneTerraformOptions, keyPath)
}

func (s *TfpSanityProvisioningTestSuite) SetupSuite() {
	testSession := session.NewSession()
	s.session = testSession

	s.client, _, s.standaloneTerraformOptions, s.terraformOptions, s.cattleConfig = ranchers.SetupRancher(s.T(), s.session, keypath.SanityKeyPath)
	s.rancherConfig, s.terraformConfig, s.terratestConfig, s.standaloneConfig = config.LoadTFPConfigs(s.cattleConfig)
}

func (s *TfpSanityProvisioningTestSuite) TestTfpProvisioningSanity() {
	var err error
	var testUser, testPassword string
	var clusterIDs []string

	s.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(s.client)
	require.NoError(s.T(), err)

	standardUserToken, err := ranchers.CreateStandardUserToken(s.T(), s.terraformOptions, s.rancherConfig, testUser, testPassword)
	require.NoError(s.T(), err)

	standardToken := standardUserToken.Token

	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}
	nodeRolesWindows := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool, config.WindowsNodePool}
	rke2Module, rke2Windows2019, rke2Windows2022, k3sModule := provisioning.DownstreamClusterModules(s.terraformConfig)

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		module    string
	}{
		{"Sanity_RKE2", nodeRolesDedicated, rke2Module},
		{"Sanity_RKE2_Windows_2019", nodeRolesWindows, rke2Windows2019},
		{"Sanity_RKE2_Windows_2022", nodeRolesWindows, rke2Windows2022},
		{"Sanity_K3S", nodeRolesDedicated, k3sModule},
	}

	for _, tt := range tests {
		if strings.Contains(tt.name, "Windows") && (s.terraformConfig.Provider != aws.Aws) {
			s.T().Skip("Skipping Windows test on non-AWS provider")
		}

		s.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(s.cattleConfig)
			rancher.AdminToken = standardToken
			terratest.Nodepools = tt.nodeRoles
			terraform.Module = tt.module

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(s.terraformConfig, s.terratestConfig, s.terraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()

			terratest, err = provisioning.GetK8sVersion(s.standardUserClient, terraform, terratest)
			require.NoError(t, err)

			terraform = provisioning.UniquifyTerraform(terraform)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, s.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(s.T(), perTestTerraformOptions, keyPath)

			logrus.Infof("Provisioning cluster (%s)", terraform.ResourcePrefix)
			clusters, customClusterName := provisioning.Provision(s.T(), s.client, s.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, newFile, rootBody, file, false, false, true, clusterIDs, "", nestedRancherModuleDir)

			logrus.Infof("Verifying the cluster is ready (%s)", clusters[0].Name)
			err = provisioningActions.VerifyClusterReady(s.client, clusters[0])
			require.NoError(s.T(), err)

			logrus.Infof("Verifying service account token secret (%s)", clusters[0].Name)
			err = clusterActions.VerifyServiceAccountTokenSecret(s.client, clusters[0].Name)
			require.NoError(s.T(), err)

			logrus.Infof("Verifying cluster pods (%s)", clusters[0].Name)
			err = pods.VerifyClusterPods(s.client, clusters[0])
			require.NoError(s.T(), err)

			if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
				logrus.Infof("Provisioning cluster (%s)", terraform.ResourcePrefix)
				clusters, _ = provisioning.Provision(s.T(), s.client, s.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, newFile, rootBody, file, true, true, true, clusterIDs, customClusterName, nestedRancherModuleDir)

				logrus.Infof("Verifying the cluster is ready (%s)", clusters[0].Name)
				err = provisioningActions.VerifyClusterReady(s.client, clusters[0])
				require.NoError(s.T(), err)

				logrus.Infof("Verifying service account token secret (%s)", clusters[0].Name)
				err = clusterActions.VerifyServiceAccountTokenSecret(s.client, clusters[0].Name)
				require.NoError(s.T(), err)

				logrus.Infof("Verifying cluster pods (%s)", clusters[0].Name)
				err = pods.VerifyClusterPods(s.client, clusters[0])
				require.NoError(s.T(), err)
			}

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

func (s *TfpSanityProvisioningTestSuite) TestTfpProvisioningSanityImported() {
	var err error
	var testUser, testPassword string
	var clusterIDs []string

	s.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(s.client)
	require.NoError(s.T(), err)

	standardUserToken, err := ranchers.CreateStandardUserToken(s.T(), s.terraformOptions, s.rancherConfig, testUser, testPassword)
	require.NoError(s.T(), err)

	standardToken := standardUserToken.Token

	rke2ImportedModule, _, _, k3sImportedModule := provisioning.ImportedClusterModules(s.terraformConfig)

	tests := []struct {
		name   string
		module string
	}{
		{"Sanity_Imported_RKE2", rke2ImportedModule},
		{"Sanity_Imported_K3S", k3sImportedModule},
	}

	for _, tt := range tests {
		if s.terraformConfig.ARMAchitecture || s.terraformConfig.MixedArchitecture {
			s.T().Skip("Skipping imported tests")
		}

		s.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(s.cattleConfig)

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(s.terraformConfig, s.terratestConfig, s.terraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()

			rancher.AdminToken = standardToken
			terraform.Module = tt.module

			terratest, err = provisioning.GetK8sVersion(s.standardUserClient, terraform, terratest)
			require.NoError(t, err)

			terraform = provisioning.UniquifyTerraform(terraform)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, s.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(s.T(), perTestTerraformOptions, keyPath)

			logrus.Infof("Provisioning cluster (%s)", terraform.ResourcePrefix)
			clusters, _ := provisioning.Provision(s.T(), s.client, s.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, newFile, rootBody, file, false, false, true, clusterIDs, "", nestedRancherModuleDir)

			logrus.Infof("Verifying the cluster is ready (%s)", clusters[0].Name)
			err = provisioningActions.VerifyClusterReady(s.client, clusters[0])
			require.NoError(s.T(), err)

			logrus.Infof("Verifying service account token secret (%s)", clusters[0].Name)
			err = clusterActions.VerifyServiceAccountTokenSecret(s.client, clusters[0].Name)
			require.NoError(s.T(), err)

			logrus.Infof("Verifying cluster pods (%s)", clusters[0].Name)
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

func TestTfpSanityProvisioningTestSuite(t *testing.T) {
	suite.Run(t, new(TfpSanityProvisioningTestSuite))
}
