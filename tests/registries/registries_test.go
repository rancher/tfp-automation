package registries

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

type TfpRegistriesTestSuite struct {
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
	authRegistry               string
	nonAuthRegistry            string
	globalRegistry             string
}

func (r *TfpRegistriesTestSuite) TearDownSuite() {
	_, keyPath := rancher2.SetKeyPath(keypath.RegistryKeyPath, r.terratestConfig.PathToRepo, r.terraformConfig.Provider)
	cleanup.Cleanup(r.T(), r.standaloneTerraformOptions, keyPath)
}

func (r *TfpRegistriesTestSuite) SetupSuite() {
	testSession := session.NewSession()
	r.session = testSession

	r.client, r.authRegistry, r.nonAuthRegistry, r.globalRegistry, r.standaloneTerraformOptions, r.terraformOptions,
		r.cattleConfig = ranchers.SetupRegistryRancher(r.T(), r.session, keypath.RegistryKeyPath)
	r.rancherConfig, r.terraformConfig, r.terratestConfig, r.standaloneConfig = config.LoadTFPConfigs(r.cattleConfig)
}

func (r *TfpRegistriesTestSuite) TestTfpAuthenticatedRegistry() {
	var err error
	var testUser, testPassword string
	var clusterIDs []string

	r.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(r.client)
	require.NoError(r.T(), err)

	standardUserToken, err := ranchers.CreateStandardUserToken(r.T(), r.terraformOptions, r.rancherConfig, testUser, testPassword)
	require.NoError(r.T(), err)

	standardToken := standardUserToken.Token

	nodeRolesAll := []config.Nodepool{config.AllRolesNodePool}
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		module    string
		nodeRoles []config.Nodepool
	}{
		{"Auth_RKE2", modules.NodeDriverAWSRKE2, nodeRolesDedicated},
		{"Auth_K3S", modules.NodeDriverAWSK3S, nodeRolesAll},
	}

	for _, tt := range tests {
		r.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(r.cattleConfig)
			rancher.AdminToken = standardToken
			terratest.Nodepools = tt.nodeRoles
			terraform.Module = tt.module
			terraform.PrivateRegistries.SystemDefaultRegistry = r.authRegistry
			terraform.PrivateRegistries.URL = r.authRegistry
			terraform.StandaloneRegistry.Authenticated = true

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(r.terraformConfig, r.terratestConfig, r.terraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()
			terratest, err = provisioning.GetK8sVersion(r.standardUserClient, terraform, terratest)
			require.NoError(r.T(), err)

			terraform = provisioning.UniquifyTerraform(terraform)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), perTestTerraformOptions, keyPath)

			logrus.Infof("Provisioning cluster (%s)", terraform.ResourcePrefix)
			clusters, _ := provisioning.Provision(r.T(), r.client, r.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, newFile, rootBody, file, false, false, true, clusterIDs, "", nestedRancherModuleDir)

			logrus.Infof("Verifying the cluster is ready (%s)", clusters[0].Name)
			err = provisioningActions.VerifyClusterReady(r.client, clusters[0])
			require.NoError(r.T(), err)

			logrus.Infof("Verifying service account token secret (%s)", clusters[0].Name)
			err = clusterActions.VerifyServiceAccountTokenSecret(r.client, clusters[0].Name)
			require.NoError(r.T(), err)

			logrus.Infof("Verifying cluster pods (%s)", clusters[0].Name)
			err = pods.VerifyClusterPods(r.client, clusters[0])
			require.NoError(r.T(), err)

			provisioning.VerifyRegistry(r.T(), r.client, clusters[0].ID, terraform)

			params := tfpQase.GetProvisioningSchemaParams(r.terraformConfig, r.terratestConfig)
			err = qase.UpdateSchemaParameters(tt.name, params)
			if err != nil {
				logrus.Warningf("Failed to upload schema parameters %s", err)
			}
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		results.ReportTest(r.terratestConfig)
	}
}

func (r *TfpRegistriesTestSuite) TestTfpECRRegistry() {
	var err error
	var testUser, testPassword string
	var clusterIDs []string

	r.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(r.client)
	require.NoError(r.T(), err)

	standardUserToken, err := ranchers.CreateStandardUserToken(r.T(), r.terraformOptions, r.rancherConfig, testUser, testPassword)
	require.NoError(r.T(), err)

	standardToken := standardUserToken.Token

	nodeRolesAll := []config.Nodepool{config.AllRolesNodePool}
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		module    string
		nodeRoles []config.Nodepool
	}{
		{"ECR_RKE2", modules.NodeDriverAWSRKE2, nodeRolesDedicated},
		{"ECR_K3S", modules.NodeDriverAWSK3S, nodeRolesAll},
	}

	for _, tt := range tests {
		r.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(r.cattleConfig)
			rancher.AdminToken = standardToken
			terratest.Nodepools = tt.nodeRoles
			terraform.Module = tt.module
			terraform.PrivateRegistries.SystemDefaultRegistry = r.terraformConfig.StandaloneRegistry.ECRURI
			terraform.PrivateRegistries.URL = r.terraformConfig.StandaloneRegistry.ECRURI
			terraform.PrivateRegistries.Username = r.terraformConfig.StandaloneRegistry.ECRUsername
			terraform.PrivateRegistries.Password = r.terraformConfig.StandaloneRegistry.ECRPassword
			terraform.StandaloneRegistry.Authenticated = true
			terraform.PrivateRegistries.AuthConfigSecretName = r.terraformConfig.PrivateRegistries.AuthConfigSecretName + "-ecr"

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(r.terraformConfig, r.terratestConfig, r.terraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()
			terratest, err = provisioning.GetK8sVersion(r.standardUserClient, terraform, terratest)
			require.NoError(r.T(), err)

			terraform = provisioning.UniquifyTerraform(terraform)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), perTestTerraformOptions, keyPath)

			logrus.Infof("Provisioning cluster (%s)", terraform.ResourcePrefix)
			clusters, _ := provisioning.Provision(r.T(), r.client, r.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, newFile, rootBody, file, false, false, true, clusterIDs, "", nestedRancherModuleDir)

			logrus.Infof("Verifying the cluster is ready (%s)", clusters[0].Name)
			err = provisioningActions.VerifyClusterReady(r.client, clusters[0])
			require.NoError(r.T(), err)

			logrus.Infof("Verifying service account token secret (%s)", clusters[0].Name)
			err = clusterActions.VerifyServiceAccountTokenSecret(r.client, clusters[0].Name)
			require.NoError(r.T(), err)

			logrus.Infof("Verifying cluster pods (%s)", clusters[0].Name)
			err = pods.VerifyClusterPods(r.client, clusters[0])
			require.NoError(r.T(), err)

			provisioning.VerifyRegistry(r.T(), r.client, clusters[0].ID, terraform)

			params := tfpQase.GetProvisioningSchemaParams(r.terraformConfig, r.terratestConfig)
			err = qase.UpdateSchemaParameters(tt.name, params)
			if err != nil {
				logrus.Warningf("Failed to upload schema parameters %s", err)
			}
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		results.ReportTest(r.terratestConfig)
	}
}

func (r *TfpRegistriesTestSuite) TestTfpGlobalRegistry() {
	var err error
	var testUser, testPassword string
	var clusterIDs []string

	nodeRolesAll := []config.Nodepool{config.AllRolesNodePool}
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	r.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(r.client)
	require.NoError(r.T(), err)

	standardUserToken, err := ranchers.CreateStandardUserToken(r.T(), r.terraformOptions, r.rancherConfig, testUser, testPassword)
	require.NoError(r.T(), err)

	standardToken := standardUserToken.Token
	nodeRolesWindows := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool, config.WindowsNodePool}

	tests := []struct {
		name      string
		module    string
		nodeRoles []config.Nodepool
	}{
		{"Global_RKE2", modules.NodeDriverAWSRKE2, nodeRolesDedicated},
		{"Global_RKE2_Windows_2019", modules.CustomAWSRKE2Windows2019, nodeRolesWindows},
		{"Global_RKE2_Windows_2022", modules.CustomAWSRKE2Windows2022, nodeRolesWindows},
		{"Global_K3S", modules.NodeDriverAWSK3S, nodeRolesAll},
	}

	for _, tt := range tests {
		r.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(r.cattleConfig)
			rancher.AdminToken = standardToken
			terratest.Nodepools = tt.nodeRoles
			terraform.Module = tt.module
			terraform.PrivateRegistries.SystemDefaultRegistry = r.globalRegistry
			terraform.PrivateRegistries.URL = r.globalRegistry
			terraform.PrivateRegistries.Password = ""
			terraform.PrivateRegistries.Username = ""
			terraform.StandaloneRegistry.Authenticated = false

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(r.terraformConfig, r.terratestConfig, r.terraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()
			terratest, err = provisioning.GetK8sVersion(r.standardUserClient, terraform, terratest)
			require.NoError(r.T(), err)

			terraform = provisioning.UniquifyTerraform(terraform)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), perTestTerraformOptions, keyPath)

			logrus.Infof("Provisioning cluster (%s)", terraform.ResourcePrefix)
			clusters, customClusterName := provisioning.Provision(r.T(), r.client, r.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, newFile, rootBody, file, false, false, true, clusterIDs, "", nestedRancherModuleDir)

			logrus.Infof("Verifying the cluster is ready (%s)", clusters[0].Name)
			err = provisioningActions.VerifyClusterReady(r.client, clusters[0])
			require.NoError(r.T(), err)

			logrus.Infof("Verifying service account token secret (%s)", clusters[0].Name)
			err = clusterActions.VerifyServiceAccountTokenSecret(r.client, clusters[0].Name)
			require.NoError(r.T(), err)

			logrus.Infof("Verifying cluster pods (%s)", clusters[0].Name)
			err = pods.VerifyClusterPods(r.client, clusters[0])
			require.NoError(r.T(), err)

			provisioning.VerifyRegistry(r.T(), r.client, clusters[0].ID, terraform)

			if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
				logrus.Infof("Provisioning cluster (%s)", terraform.ResourcePrefix)
				clusters, _ = provisioning.Provision(r.T(), r.client, r.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, newFile, rootBody, file, true, true, true, clusterIDs, customClusterName, nestedRancherModuleDir)

				logrus.Infof("Verifying the cluster is ready (%s)", clusters[0].Name)
				err = provisioningActions.VerifyClusterReady(r.client, clusters[0])
				require.NoError(r.T(), err)

				logrus.Infof("Verifying service account token secret (%s)", clusters[0].Name)
				err = clusterActions.VerifyServiceAccountTokenSecret(r.client, clusters[0].Name)
				require.NoError(r.T(), err)

				logrus.Infof("Verifying cluster pods (%s)", clusters[0].Name)
				err = pods.VerifyClusterPods(r.client, clusters[0])
				require.NoError(r.T(), err)

				provisioning.VerifyRegistry(r.T(), r.client, clusters[0].ID, terraform)
			}

			params := tfpQase.GetProvisioningSchemaParams(r.terraformConfig, r.terratestConfig)
			err = qase.UpdateSchemaParameters(tt.name, params)
			if err != nil {
				logrus.Warningf("Failed to upload schema parameters %s", err)
			}
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		results.ReportTest(r.terratestConfig)
	}
}

func (r *TfpRegistriesTestSuite) TestTfpNonAuthenticatedRegistry() {
	var err error
	var testUser, testPassword string
	var clusterIDs []string

	r.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(r.client)
	require.NoError(r.T(), err)

	standardUserToken, err := ranchers.CreateStandardUserToken(r.T(), r.terraformOptions, r.rancherConfig, testUser, testPassword)
	require.NoError(r.T(), err)

	standardToken := standardUserToken.Token

	nodeRolesAll := []config.Nodepool{config.AllRolesNodePool}
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}
	nodeRolesWindows := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool, config.WindowsNodePool}

	tests := []struct {
		name      string
		module    string
		nodeRoles []config.Nodepool
	}{
		{"Non_Auth_RKE2", modules.NodeDriverAWSRKE2, nodeRolesDedicated},
		{"Non_Auth_RKE2_Windows_2019", modules.CustomAWSRKE2Windows2019, nodeRolesWindows},
		{"Non_Auth_RKE2_Windows_2022", modules.CustomAWSRKE2Windows2022, nodeRolesWindows},
		{"Non_Auth_K3S", modules.NodeDriverAWSK3S, nodeRolesAll},
	}

	for _, tt := range tests {
		r.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(r.cattleConfig)
			rancher.AdminToken = standardToken
			terratest.Nodepools = tt.nodeRoles
			terraform.Module = tt.module
			terraform.PrivateRegistries.SystemDefaultRegistry = r.nonAuthRegistry
			terraform.PrivateRegistries.URL = r.nonAuthRegistry
			terraform.PrivateRegistries.Password = ""
			terraform.PrivateRegistries.Username = ""
			terraform.StandaloneRegistry.Authenticated = false

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(r.terraformConfig, r.terratestConfig, r.terraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()
			terratest, err = provisioning.GetK8sVersion(r.standardUserClient, terraform, terratest)
			require.NoError(r.T(), err)

			terraform = provisioning.UniquifyTerraform(terraform)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), perTestTerraformOptions, keyPath)

			logrus.Infof("Provisioning cluster (%s)", terraform.ResourcePrefix)
			clusters, customClusterName := provisioning.Provision(r.T(), r.client, r.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, newFile, rootBody, file, false, false, true, clusterIDs, "", nestedRancherModuleDir)

			logrus.Infof("Verifying the cluster is ready (%s)", clusters[0].Name)
			err = provisioningActions.VerifyClusterReady(r.client, clusters[0])
			require.NoError(r.T(), err)

			logrus.Infof("Verifying service account token secret (%s)", clusters[0].Name)
			err = clusterActions.VerifyServiceAccountTokenSecret(r.client, clusters[0].Name)
			require.NoError(r.T(), err)

			logrus.Infof("Verifying cluster pods (%s)", clusters[0].Name)
			err = pods.VerifyClusterPods(r.client, clusters[0])
			require.NoError(r.T(), err)

			provisioning.VerifyRegistry(r.T(), r.client, clusters[0].ID, terraform)

			if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
				logrus.Infof("Provisioning cluster (%s)", terraform.ResourcePrefix)
				clusters, _ = provisioning.Provision(r.T(), r.client, r.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, newFile, rootBody, file, true, true, true, clusterIDs, customClusterName, nestedRancherModuleDir)

				logrus.Infof("Verifying the cluster is ready (%s)", clusters[0].Name)
				err = provisioningActions.VerifyClusterReady(r.client, clusters[0])
				require.NoError(r.T(), err)

				logrus.Infof("Verifying service account token secret (%s)", clusters[0].Name)
				err = clusterActions.VerifyServiceAccountTokenSecret(r.client, clusters[0].Name)
				require.NoError(r.T(), err)

				logrus.Infof("Verifying cluster pods (%s)", clusters[0].Name)
				err = pods.VerifyClusterPods(r.client, clusters[0])
				require.NoError(r.T(), err)

				provisioning.VerifyRegistry(r.T(), r.client, clusters[0].ID, terraform)
			}

			params := tfpQase.GetProvisioningSchemaParams(r.terraformConfig, r.terratestConfig)
			err = qase.UpdateSchemaParameters(tt.name, params)
			if err != nil {
				logrus.Warningf("Failed to upload schema parameters %s", err)
			}
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		results.ReportTest(r.terratestConfig)
	}
}

func TestTfpRegistriesTestSuite(t *testing.T) {
	suite.Run(t, new(TfpRegistriesTestSuite))
}
