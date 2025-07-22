package registries

import (
	"os"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/registries"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/infrastructure"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TfpRegistriesTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	session                    *session.Session
	cattleConfig               map[string]any
	rancherConfig              *rancher.Config
	terraformConfig            *config.TerraformConfig
	terratestConfig            *config.TerratestConfig
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
	r.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	r.rancherConfig, r.terraformConfig, r.terratestConfig = config.LoadTFPConfigs(r.cattleConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.RegistryKeyPath, r.terratestConfig.PathToRepo, r.terraformConfig.Provider)
	standaloneTerraformOptions := framework.Setup(r.T(), r.terraformConfig, r.terratestConfig, keyPath)
	r.standaloneTerraformOptions = standaloneTerraformOptions

	authRegistry, nonAuthRegistry, globalRegistry, err := registries.CreateMainTF(r.T(), r.standaloneTerraformOptions, keyPath, r.rancherConfig, r.terraformConfig, r.terratestConfig)
	require.NoError(r.T(), err)

	r.authRegistry = authRegistry
	r.nonAuthRegistry = nonAuthRegistry
	r.globalRegistry = globalRegistry

	testSession := session.NewSession()
	r.session = testSession

	client, err := infrastructure.PostRancherSetup(r.T(), r.rancherConfig, testSession, r.terraformConfig.Standalone.RancherHostname, false, false)
	require.NoError(r.T(), err)

	r.client = client

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(r.T(), r.terraformConfig, r.terratestConfig, keyPath)
	r.terraformOptions = terraformOptions
}

func (r *TfpRegistriesTestSuite) TestTfpGlobalRegistry() {
	nodeRolesAll := []config.Nodepool{config.AllRolesNodePool}
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		module    string
		nodeRoles []config.Nodepool
	}{
		{"Global RKE2", modules.EC2RKE2, nodeRolesDedicated},
		{"Global K3S", modules.EC2K3s, nodeRolesAll},
	}

	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(r.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{r.cattleConfig})
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, r.client.RancherConfig.AdminToken, configMap[0])
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

		provisioning.GetK8sVersion(r.T(), r.client, r.terratestConfig, r.terraformConfig, configs.DefaultK8sVersion, configMap)

		rancher, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		currentDate := time.Now().Format("2006-01-02 03:04PM")
		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion + " " + currentDate

		r.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			clusterIDs, _ := provisioning.Provision(r.T(), r.client, rancher, terraform, terratest, testUser, testPassword, r.terraformOptions, configMap, newFile, rootBody, file, false, false, true, nil)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], terraform)
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		qase.ReportTest(r.terratestConfig)
	}
}

func (r *TfpRegistriesTestSuite) TestTfpAuthenticatedRegistry() {
	nodeRolesAll := []config.Nodepool{config.AllRolesNodePool}
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		module    string
		nodeRoles []config.Nodepool
	}{
		{"Auth RKE2", modules.EC2RKE2, nodeRolesDedicated},
		{"Auth K3S", modules.EC2K3s, nodeRolesAll},
	}

	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(r.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{r.cattleConfig})
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, r.client.RancherConfig.AdminToken, configMap[0])
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

		provisioning.GetK8sVersion(r.T(), r.client, r.terratestConfig, r.terraformConfig, configs.DefaultK8sVersion, configMap)

		rancher, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		currentDate := time.Now().Format("2006-01-02 03:04PM")
		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion + " " + currentDate

		r.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			clusterIDs, _ := provisioning.Provision(r.T(), r.client, rancher, terraform, terratest, testUser, testPassword, r.terraformOptions, configMap, newFile, rootBody, file, false, false, true, nil)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], terraform)
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		qase.ReportTest(r.terratestConfig)
	}
}

func (r *TfpRegistriesTestSuite) TestTfpNonAuthenticatedRegistry() {
	nodeRolesAll := []config.Nodepool{config.AllRolesNodePool}
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		module    string
		nodeRoles []config.Nodepool
	}{
		{"Non Auth RKE2", modules.EC2RKE2, nodeRolesDedicated},
		{"Non Auth K3S", modules.EC2K3s, nodeRolesAll},
	}

	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(r.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{r.cattleConfig})
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, r.client.RancherConfig.AdminToken, configMap[0])
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

		provisioning.GetK8sVersion(r.T(), r.client, r.terratestConfig, r.terraformConfig, configs.DefaultK8sVersion, configMap)

		rancher, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		currentDate := time.Now().Format("2006-01-02 03:04PM")
		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion + " " + currentDate

		r.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			clusterIDs, _ := provisioning.Provision(r.T(), r.client, rancher, terraform, terratest, testUser, testPassword, r.terraformOptions, configMap, newFile, rootBody, file, false, false, true, nil)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], terraform)
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		qase.ReportTest(r.terratestConfig)
	}
}

func (r *TfpRegistriesTestSuite) TestTfpECRRegistry() {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		module    string
		nodeRoles []config.Nodepool
	}{
		{"ECR RKE2", "ec2_rke2", nodeRolesDedicated},
	}

	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(r.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{r.cattleConfig})
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, r.client.RancherConfig.AdminToken, configMap[0])
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

		provisioning.GetK8sVersion(r.T(), r.client, r.terratestConfig, r.terraformConfig, configs.DefaultK8sVersion, configMap)

		rancher, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		currentDate := time.Now().Format("2006-01-02 03:04PM")
		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion + " " + currentDate

		r.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			clusterIDs, _ := provisioning.Provision(r.T(), r.client, rancher, terraform, terratest, testUser, testPassword, r.terraformOptions, configMap, newFile, rootBody, file, false, false, true, nil)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], terraform)
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		qase.ReportTest(r.terratestConfig)
	}
}

func TestTfpRegistriesTestSuite(t *testing.T) {
	suite.Run(t, new(TfpRegistriesTestSuite))
}
