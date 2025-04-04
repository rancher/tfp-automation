package registries

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/token"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/pipeline"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/registries"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
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
	keyPath := rancher2.SetKeyPath(keypath.RegistryKeyPath, r.terraformConfig.Provider)
	cleanup.Cleanup(r.T(), r.standaloneTerraformOptions, keyPath)
}

func (r *TfpRegistriesTestSuite) SetupSuite() {
	r.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	r.rancherConfig, r.terraformConfig, r.terratestConfig = config.LoadTFPConfigs(r.cattleConfig)

	keyPath := rancher2.SetKeyPath(keypath.RegistryKeyPath, r.terraformConfig.Provider)
	standaloneTerraformOptions := framework.Setup(r.T(), r.terraformConfig, r.terratestConfig, keyPath)
	r.standaloneTerraformOptions = standaloneTerraformOptions

	authRegistry, nonAuthRegistry, globalRegistry, err := registries.CreateMainTF(r.T(), r.standaloneTerraformOptions, keyPath, r.terraformConfig, r.terratestConfig)
	require.NoError(r.T(), err)

	r.authRegistry = authRegistry
	r.nonAuthRegistry = nonAuthRegistry
	r.globalRegistry = globalRegistry
}

func (r *TfpRegistriesTestSuite) TfpSetupSuite() map[string]any {
	testSession := session.NewSession()
	r.session = testSession

	r.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	configMap, err := provisioning.UniquifyTerraform([]map[string]any{r.cattleConfig})
	require.NoError(r.T(), err)

	r.cattleConfig = configMap[0]
	r.rancherConfig, r.terraformConfig, r.terratestConfig = config.LoadTFPConfigs(r.cattleConfig)

	adminUser := &management.User{
		Username: "admin",
		Password: r.rancherConfig.AdminPassword,
	}

	userToken, err := token.GenerateUserToken(adminUser, r.rancherConfig.Host)
	require.NoError(r.T(), err)

	r.rancherConfig.AdminToken = userToken.Token

	client, err := rancher.NewClient(r.rancherConfig.AdminToken, testSession)
	require.NoError(r.T(), err)

	r.client = client
	r.client.RancherConfig.AdminToken = r.rancherConfig.AdminToken
	r.client.RancherConfig.AdminPassword = r.rancherConfig.AdminPassword
	r.client.RancherConfig.Host = r.rancherConfig.Host

	operations.ReplaceValue([]string{"rancher", "adminToken"}, r.rancherConfig.AdminToken, configMap[0])
	operations.ReplaceValue([]string{"rancher", "adminPassword"}, r.rancherConfig.AdminPassword, configMap[0])
	operations.ReplaceValue([]string{"rancher", "host"}, r.rancherConfig.Host, configMap[0])

	err = pipeline.PostRancherInstall(r.client, r.client.RancherConfig.AdminPassword)
	require.NoError(r.T(), err)

	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, "")
	terraformOptions := framework.Setup(r.T(), r.terraformConfig, r.terratestConfig, keyPath)
	r.terraformOptions = terraformOptions

	return r.cattleConfig
}

func (r *TfpRegistriesTestSuite) TestTfpGlobalRegistry() {
	nodeRolesAll := config.AllRolesNodePool
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		module    string
		nodeRoles []config.Nodepool
	}{
		{"Global RKE1", "ec2_rke1", nodeRolesDedicated},
		{"Global RKE2", "ec2_rke2", nodeRolesDedicated},
		{"Global K3S", "ec2_k3s", []config.Nodepool{nodeRolesAll}},
	}

	for _, tt := range tests {
		cattleConfig := r.TfpSetupSuite()
		configMap := []map[string]any{cattleConfig}

		operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, r.globalRegistry, configMap[0])
		operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, r.globalRegistry, configMap[0])
		operations.ReplaceValue([]string{"terraform", "privateRegistries", "password"}, "", configMap[0])
		operations.ReplaceValue([]string{"terraform", "privateRegistries", "username"}, "", configMap[0])
		operations.ReplaceValue([]string{"terraform", "standaloneRegistry", "authenticated"}, false, configMap[0])

		provisioning.GetK8sVersion(r.T(), r.client, r.terratestConfig, r.terraformConfig, configs.DefaultK8sVersion, configMap)

		_, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion
		testUser, testPassword := configs.CreateTestCredentials()

		r.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, "")
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			clusterIDs := provisioning.Provision(r.T(), r.client, r.rancherConfig, terraform, terratest, testUser, testPassword, r.terraformOptions, configMap, false)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], terraform)
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func (r *TfpRegistriesTestSuite) TestTfpAuthenticatedRegistry() {
	nodeRolesAll := config.AllRolesNodePool
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		module    string
		nodeRoles []config.Nodepool
	}{
		{"Auth RKE1", "ec2_rke1", nodeRolesDedicated},
		{"Auth RKE2", "ec2_rke2", nodeRolesDedicated},
		{"Auth K3S", "ec2_k3s", []config.Nodepool{nodeRolesAll}},
	}

	for _, tt := range tests {
		cattleConfig := r.TfpSetupSuite()
		configMap := []map[string]any{cattleConfig}

		operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, r.authRegistry, configMap[0])
		operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, r.authRegistry, configMap[0])
		operations.ReplaceValue([]string{"terraform", "standaloneRegistry", "authenticated"}, true, configMap[0])

		provisioning.GetK8sVersion(r.T(), r.client, r.terratestConfig, r.terraformConfig, configs.DefaultK8sVersion, configMap)

		_, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion
		testUser, testPassword := configs.CreateTestCredentials()

		r.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, "")
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			clusterIDs := provisioning.Provision(r.T(), r.client, r.rancherConfig, terraform, terratest, testUser, testPassword, r.terraformOptions, configMap, false)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], terraform)
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func (r *TfpRegistriesTestSuite) TestTfpNonAuthenticatedRegistry() {
	nodeRolesAll := config.AllRolesNodePool
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		module    string
		nodeRoles []config.Nodepool
	}{
		{"Non Auth RKE1", "ec2_rke1", nodeRolesDedicated},
		{"Non Auth RKE2", "ec2_rke2", nodeRolesDedicated},
		{"Non Auth K3S", "ec2_k3s", []config.Nodepool{nodeRolesAll}},
	}

	for _, tt := range tests {
		cattleConfig := r.TfpSetupSuite()
		configMap := []map[string]any{cattleConfig}

		operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, r.nonAuthRegistry, configMap[0])
		operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, r.nonAuthRegistry, configMap[0])
		operations.ReplaceValue([]string{"terraform", "privateRegistries", "password"}, "", configMap[0])
		operations.ReplaceValue([]string{"terraform", "privateRegistries", "username"}, "", configMap[0])
		operations.ReplaceValue([]string{"terraform", "standaloneRegistry", "authenticated"}, false, configMap[0])

		provisioning.GetK8sVersion(r.T(), r.client, r.terratestConfig, r.terraformConfig, configs.DefaultK8sVersion, configMap)

		_, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion
		testUser, testPassword := configs.CreateTestCredentials()

		r.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, "")
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			clusterIDs := provisioning.Provision(r.T(), r.client, r.rancherConfig, terraform, terratest, testUser, testPassword, r.terraformOptions, configMap, false)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], terraform)
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpRegistriesTestSuite(t *testing.T) {
	suite.Run(t, new(TfpRegistriesTestSuite))
}
