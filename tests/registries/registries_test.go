package registries

import (
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
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

func (r *TfpRegistriesTestSuite) TestTfpGlobalRegistry() {
	var err error
	var testUser, testPassword string
	var clusterIDs []string

	nodeRolesAll := []config.Nodepool{config.AllRolesNodePool}
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}
	customClusterNames := []string{}

	r.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(r.client)
	require.NoError(r.T(), err)

	standardUserToken, err := ranchers.CreateStandardUserToken(r.T(), r.terraformOptions, r.rancherConfig, testUser, testPassword)
	require.NoError(r.T(), err)

	standardToken := standardUserToken.Token

	tests := []struct {
		name      string
		module    string
		nodeRoles []config.Nodepool
	}{
		{"Global_RKE2", modules.EC2RKE2, nodeRolesDedicated},
		{"Global_RKE2_Windows_2019", modules.CustomEC2RKE2Windows2019, nil},
		{"Global_RKE2_Windows_2022", modules.CustomEC2RKE2Windows2022, nil},
		{"Global_K3S", modules.EC2K3s, nodeRolesAll},
	}

	for _, tt := range tests {
		r.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(r.terraformConfig, r.terratestConfig, r.terraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()

			cattleConfig, err := provisioning.UniquifyTerraform(r.cattleConfig)
			require.NoError(t, err)

			_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, r.globalRegistry, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, r.globalRegistry, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "password"}, "", cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "username"}, "", cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "standaloneRegistry", "authenticated"}, false, cattleConfig)
			require.NoError(r.T(), err)

			provisioning.GetK8sVersion(r.standardUserClient, cattleConfig)

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(cattleConfig)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), perTestTerraformOptions, keyPath)

			clusterIDs, customClusterNames := provisioning.Provision(r.T(), r.client, r.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, newFile, rootBody, file, false, false, true, clusterIDs, customClusterNames, nestedRancherModuleDir)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			provisioning.VerifyServiceAccountTokenSecret(r.T(), r.client, clusterIDs)

			cluster, err := r.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraform.ResourcePrefix)
			require.NoError(r.T(), err)

			err = pods.VerifyClusterPods(r.client, cluster)
			require.NoError(r.T(), err)

			provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], terraform)

			if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
				clusterIDs, _ = provisioning.Provision(r.T(), r.client, r.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, newFile, rootBody, file, true, true, true, clusterIDs, customClusterNames, nestedRancherModuleDir)
				provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
				provisioning.VerifyServiceAccountTokenSecret(r.T(), r.client, clusterIDs)

				err = pods.VerifyClusterPods(r.client, cluster)
				require.NoError(r.T(), err)

				provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], terraform)
			}

			params := tfpQase.GetProvisioningSchemaParams(cattleConfig)
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
		{"Auth_RKE2", modules.EC2RKE2, nodeRolesDedicated},
		{"Auth_K3S", modules.EC2K3s, nodeRolesAll},
	}

	for _, tt := range tests {
		r.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(r.terraformConfig, r.terratestConfig, r.terraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()

			cattleConfig, err := provisioning.UniquifyTerraform(r.cattleConfig)
			require.NoError(t, err)

			_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, r.authRegistry, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, r.authRegistry, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "standaloneRegistry", "authenticated"}, true, cattleConfig)
			require.NoError(r.T(), err)

			provisioning.GetK8sVersion(r.standardUserClient, cattleConfig)

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(cattleConfig)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), perTestTerraformOptions, keyPath)

			clusterIDs, _ := provisioning.Provision(r.T(), r.client, r.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, newFile, rootBody, file, false, false, true, clusterIDs, nil, nestedRancherModuleDir)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			provisioning.VerifyServiceAccountTokenSecret(r.T(), r.client, clusterIDs)

			cluster, err := r.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraform.ResourcePrefix)
			require.NoError(r.T(), err)

			err = pods.VerifyClusterPods(r.client, cluster)
			require.NoError(r.T(), err)

			provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], terraform)

			params := tfpQase.GetProvisioningSchemaParams(cattleConfig)
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
	customClusterNames := []string{}

	tests := []struct {
		name      string
		module    string
		nodeRoles []config.Nodepool
	}{
		{"Non_Auth_RKE2", modules.EC2RKE2, nodeRolesDedicated},
		{"Non_Auth_RKE2_Windows_2019", modules.CustomEC2RKE2Windows2019, nil},
		{"Non_Auth_RKE2_Windows_2022", modules.CustomEC2RKE2Windows2022, nil},
		{"Non_Auth_K3S", modules.EC2K3s, nodeRolesAll},
	}

	for _, tt := range tests {
		r.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(r.terraformConfig, r.terratestConfig, r.terraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()

			cattleConfig, err := provisioning.UniquifyTerraform(r.cattleConfig)
			require.NoError(t, err)

			_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, r.nonAuthRegistry, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, r.nonAuthRegistry, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "password"}, "", cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "username"}, "", cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "standaloneRegistry", "authenticated"}, false, cattleConfig)
			require.NoError(r.T(), err)

			provisioning.GetK8sVersion(r.standardUserClient, cattleConfig)

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(cattleConfig)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), perTestTerraformOptions, keyPath)

			clusterIDs, customClusterNames := provisioning.Provision(r.T(), r.client, r.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, newFile, rootBody, file, false, false, true, clusterIDs, customClusterNames, nestedRancherModuleDir)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			provisioning.VerifyServiceAccountTokenSecret(r.T(), r.client, clusterIDs)

			cluster, err := r.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraform.ResourcePrefix)
			require.NoError(r.T(), err)

			err = pods.VerifyClusterPods(r.client, cluster)
			require.NoError(r.T(), err)

			provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], terraform)

			if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
				clusterIDs, _ = provisioning.Provision(r.T(), r.client, r.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, newFile, rootBody, file, true, true, true, clusterIDs, customClusterNames, nestedRancherModuleDir)
				provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
				provisioning.VerifyServiceAccountTokenSecret(r.T(), r.client, clusterIDs)

				err = pods.VerifyClusterPods(r.client, cluster)
				require.NoError(r.T(), err)

				provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], terraform)
			}

			params := tfpQase.GetProvisioningSchemaParams(cattleConfig)
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
		{"ECR_RKE2", modules.EC2RKE2, nodeRolesDedicated},
		{"ECR_K3S", modules.EC2K3s, nodeRolesAll},
	}

	for _, tt := range tests {
		r.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(r.terraformConfig, r.terratestConfig, r.terraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()

			cattleConfig, err := provisioning.UniquifyTerraform(r.cattleConfig)
			require.NoError(t, err)

			_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, r.terraformConfig.StandaloneRegistry.ECRURI, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, r.terraformConfig.StandaloneRegistry.ECRURI, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "username"}, r.terraformConfig.StandaloneRegistry.ECRUsername, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "password"}, r.terraformConfig.StandaloneRegistry.ECRPassword, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "standaloneRegistry", "authenticated"}, true, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "authConfigSecretName"}, r.terraformConfig.PrivateRegistries.AuthConfigSecretName+"-ecr", cattleConfig)
			require.NoError(r.T(), err)

			provisioning.GetK8sVersion(r.standardUserClient, cattleConfig)

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(cattleConfig)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), perTestTerraformOptions, keyPath)

			clusterIDs, _ := provisioning.Provision(r.T(), r.client, r.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, newFile, rootBody, file, false, false, true, clusterIDs, nil, nestedRancherModuleDir)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			provisioning.VerifyServiceAccountTokenSecret(r.T(), r.client, clusterIDs)

			cluster, err := r.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraform.ResourcePrefix)
			require.NoError(r.T(), err)

			err = pods.VerifyClusterPods(r.client, cluster)
			require.NoError(r.T(), err)

			provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], terraform)

			params := tfpQase.GetProvisioningSchemaParams(cattleConfig)
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
