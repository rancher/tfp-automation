package registries

import (
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
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
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

	nodeRolesAll := []config.Nodepool{config.AllRolesNodePool}
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

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
		{"Global_K3S", modules.EC2K3s, nodeRolesAll},
	}

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(r.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{r.cattleConfig})
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, r.globalRegistry, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, r.globalRegistry, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "password"}, "", configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "username"}, "", configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "standaloneRegistry", "authenticated"}, false, configMap[0])
		require.NoError(r.T(), err)

		provisioning.GetK8sVersion(r.T(), r.standardUserClient, r.terratestConfig, r.terraformConfig, configMap)

		rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

		r.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			clusterIDs, _ := provisioning.Provision(r.T(), r.client, r.standardUserClient, rancher, terraform, terratest, testUser, testPassword, r.terraformOptions, configMap, newFile, rootBody, file, false, false, true, nil)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			provisioning.VerifyServiceAccountTokenSecret(r.T(), r.client, clusterIDs)

			cluster, err := r.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraform.ResourcePrefix)
			require.NoError(r.T(), err)

			err = pods.VerifyClusterPods(r.client, cluster)
			require.NoError(r.T(), err)

			provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], terraform)
		})

		params := tfpQase.GetProvisioningSchemaParams(configMap[0])
		err = qase.UpdateSchemaParameters(tt.name, params)
		if err != nil {
			logrus.Warningf("Failed to upload schema parameters %s", err)
		}
	}

	if r.terratestConfig.LocalQaseReporting {
		results.ReportTest(r.terratestConfig)
	}
}

func (r *TfpRegistriesTestSuite) TestTfpAuthenticatedRegistry() {
	var err error
	var testUser, testPassword string

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
		newFile, rootBody, file := rancher2.InitializeMainTF(r.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{r.cattleConfig})
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, r.authRegistry, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, r.authRegistry, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "standaloneRegistry", "authenticated"}, true, configMap[0])
		require.NoError(r.T(), err)

		provisioning.GetK8sVersion(r.T(), r.standardUserClient, r.terratestConfig, r.terraformConfig, configMap)

		rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

		r.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			clusterIDs, _ := provisioning.Provision(r.T(), r.client, r.standardUserClient, rancher, terraform, terratest, testUser, testPassword, r.terraformOptions, configMap, newFile, rootBody, file, false, false, true, nil)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			provisioning.VerifyServiceAccountTokenSecret(r.T(), r.client, clusterIDs)

			cluster, err := r.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraform.ResourcePrefix)
			require.NoError(r.T(), err)

			err = pods.VerifyClusterPods(r.client, cluster)
			require.NoError(r.T(), err)

			provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], terraform)
		})

		params := tfpQase.GetProvisioningSchemaParams(configMap[0])
		err = qase.UpdateSchemaParameters(tt.name, params)
		if err != nil {
			logrus.Warningf("Failed to upload schema parameters %s", err)
		}
	}

	if r.terratestConfig.LocalQaseReporting {
		results.ReportTest(r.terratestConfig)
	}
}

func (r *TfpRegistriesTestSuite) TestTfpNonAuthenticatedRegistry() {
	var err error
	var testUser, testPassword string

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
		{"Non_Auth_RKE2", modules.EC2RKE2, nodeRolesDedicated},
		{"Non_Auth_K3S", modules.EC2K3s, nodeRolesAll},
	}

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(r.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{r.cattleConfig})
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, r.nonAuthRegistry, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, r.nonAuthRegistry, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "password"}, "", configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "username"}, "", configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "standaloneRegistry", "authenticated"}, false, configMap[0])
		require.NoError(r.T(), err)

		provisioning.GetK8sVersion(r.T(), r.standardUserClient, r.terratestConfig, r.terraformConfig, configMap)

		rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

		r.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			clusterIDs, _ := provisioning.Provision(r.T(), r.client, r.standardUserClient, rancher, terraform, terratest, testUser, testPassword, r.terraformOptions, configMap, newFile, rootBody, file, false, false, true, nil)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			provisioning.VerifyServiceAccountTokenSecret(r.T(), r.client, clusterIDs)

			cluster, err := r.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraform.ResourcePrefix)
			require.NoError(r.T(), err)

			err = pods.VerifyClusterPods(r.client, cluster)
			require.NoError(r.T(), err)

			provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], terraform)
		})

		params := tfpQase.GetProvisioningSchemaParams(configMap[0])
		err = qase.UpdateSchemaParameters(tt.name, params)
		if err != nil {
			logrus.Warningf("Failed to upload schema parameters %s", err)
		}
	}

	if r.terratestConfig.LocalQaseReporting {
		results.ReportTest(r.terratestConfig)
	}
}

func (r *TfpRegistriesTestSuite) TestTfpECRRegistry() {
	var err error
	var testUser, testPassword string

	r.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(r.client)
	require.NoError(r.T(), err)

	standardUserToken, err := ranchers.CreateStandardUserToken(r.T(), r.terraformOptions, r.rancherConfig, testUser, testPassword)
	require.NoError(r.T(), err)

	standardToken := standardUserToken.Token

	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		module    string
		nodeRoles []config.Nodepool
	}{
		{"ECR_RKE2", "ec2_rke2", nodeRolesDedicated},
	}

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(r.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{r.cattleConfig})
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, r.terraformConfig.StandaloneRegistry.ECRURI, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, r.terraformConfig.StandaloneRegistry.ECRURI, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "username"}, r.terraformConfig.StandaloneRegistry.ECRUsername, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "password"}, r.terraformConfig.StandaloneRegistry.ECRPassword, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "standaloneRegistry", "authenticated"}, true, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "authConfigSecretName"}, r.terraformConfig.PrivateRegistries.AuthConfigSecretName+"-ecr", configMap[0])
		require.NoError(r.T(), err)

		provisioning.GetK8sVersion(r.T(), r.standardUserClient, r.terratestConfig, r.terraformConfig, configMap)

		rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

		r.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			clusterIDs, _ := provisioning.Provision(r.T(), r.client, r.standardUserClient, rancher, terraform, terratest, testUser, testPassword, r.terraformOptions, configMap, newFile, rootBody, file, false, false, true, nil)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			provisioning.VerifyServiceAccountTokenSecret(r.T(), r.client, clusterIDs)

			cluster, err := r.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraform.ResourcePrefix)
			require.NoError(r.T(), err)

			err = pods.VerifyClusterPods(r.client, cluster)
			require.NoError(r.T(), err)

			provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], terraform)
		})

		params := tfpQase.GetProvisioningSchemaParams(configMap[0])
		err = qase.UpdateSchemaParameters(tt.name, params)
		if err != nil {
			logrus.Warningf("Failed to upload schema parameters %s", err)
		}
	}

	if r.terratestConfig.LocalQaseReporting {
		results.ReportTest(r.terratestConfig)
	}
}

func TestTfpRegistriesTestSuite(t *testing.T) {
	suite.Run(t, new(TfpRegistriesTestSuite))
}
