package airgap

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/rancher/tests/v2/actions/pipeline"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/token"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/airgap"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TfpAirgapProvisioningTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	session                    *session.Session
	cattleConfig               map[string]any
	rancherConfig              *rancher.Config
	terraformConfig            *config.TerraformConfig
	terratestConfig            *config.TerratestConfig
	standaloneTerraformOptions *terraform.Options
	terraformOptions           *terraform.Options
	registry                   string
}

func (a *TfpAirgapProvisioningTestSuite) TearDownSuite() {
	keyPath := rancher2.SetKeyPath(keypath.AirgapKeyPath)
	cleanup.Cleanup(a.T(), a.standaloneTerraformOptions, keyPath)
}

func (a *TfpAirgapProvisioningTestSuite) SetupSuite() {
	a.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	a.rancherConfig, a.terraformConfig, a.terratestConfig = config.LoadTFPConfigs(a.cattleConfig)

	keyPath := rancher2.SetKeyPath(keypath.AirgapKeyPath)
	standaloneTerraformOptions := framework.Setup(a.T(), a.terraformConfig, a.terratestConfig, keyPath)
	a.standaloneTerraformOptions = standaloneTerraformOptions

	registry, err := airgap.CreateMainTF(a.T(), a.standaloneTerraformOptions, keyPath, a.terraformConfig, a.terratestConfig)
	require.NoError(a.T(), err)

	a.registry = registry
}

func (a *TfpAirgapProvisioningTestSuite) TfpSetupSuite() map[string]any {
	testSession := session.NewSession()
	a.session = testSession

	a.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	a.rancherConfig, a.terraformConfig, a.terratestConfig = config.LoadTFPConfigs(a.cattleConfig)

	adminUser := &management.User{
		Username: "admin",
		Password: a.rancherConfig.AdminPassword,
	}

	userToken, err := token.GenerateUserToken(adminUser, a.rancherConfig.Host)
	require.NoError(a.T(), err)

	a.rancherConfig.AdminToken = userToken.Token

	client, err := rancher.NewClient(a.rancherConfig.AdminToken, testSession)
	require.NoError(a.T(), err)

	a.client = client
	a.client.RancherConfig.AdminToken = a.rancherConfig.AdminToken
	a.client.RancherConfig.AdminPassword = a.rancherConfig.AdminPassword
	a.client.RancherConfig.Host = a.terraformConfig.Standalone.AirgapInternalFQDN

	err = pipeline.PostRancherInstall(a.client, a.client.RancherConfig.AdminPassword)
	require.NoError(a.T(), err)

	a.client.RancherConfig.Host = a.rancherConfig.Host

	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
	terraformOptions := framework.Setup(a.T(), a.terraformConfig, a.terratestConfig, keyPath)
	a.terraformOptions = terraformOptions

	return a.cattleConfig
}

func (a *TfpAirgapProvisioningTestSuite) TestTfpAirgapProvisioning() {
	tests := []struct {
		name   string
		module string
	}{
		{"Airgap RKE1", "airgap_rke1"},
		{"Airgap RKE2", "airgap_rke2"},
		{"Airgap K3S", "airgap_k3s"},
	}

	for _, tt := range tests {
		cattleConfig := a.TfpSetupSuite()
		configMap := []map[string]any{cattleConfig}

		operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, a.registry, configMap[0])
		operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, a.registry, configMap[0])

		provisioning.GetK8sVersion(a.T(), a.client, a.terratestConfig, a.terraformConfig, configs.DefaultK8sVersion, configMap)

		terraform := new(config.TerraformConfig)
		operations.LoadObjectFromMap(config.TerraformConfigurationFileKey, configMap[0], terraform)

		terratest := new(config.TerratestConfig)
		operations.LoadObjectFromMap(config.TerratestConfigurationFileKey, configMap[0], terratest)

		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion
		testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

		a.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
			defer cleanup.Cleanup(a.T(), a.terraformOptions, keyPath)

			clusterIDs := provisioning.Provision(a.T(), a.client, a.rancherConfig, terraform, terratest, testUser, testPassword, clusterName, poolName, a.terraformOptions, configMap)
			provisioning.VerifyClustersState(a.T(), a.client, clusterIDs)
		})
	}

	if a.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func (a *TfpAirgapProvisioningTestSuite) TestTfpAirgapUpgrading() {
	tests := []struct {
		name   string
		module string
	}{
		{"Upgrading Airgap RKE1", "airgap_rke1"},
		{"Upgrading Airgap RKE2", "airgap_rke2"},
		{"Upgrading Airgap K3S", "airgap_k3s"},
	}

	for _, tt := range tests {
		cattleConfig := a.TfpSetupSuite()
		configMap := []map[string]any{cattleConfig}

		operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, a.registry, configMap[0])
		operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, a.registry, configMap[0])

		provisioning.GetK8sVersion(a.T(), a.client, a.terratestConfig, a.terraformConfig, configs.SecondHighestVersion, configMap)

		terraform := new(config.TerraformConfig)
		operations.LoadObjectFromMap(config.TerraformConfigurationFileKey, configMap[0], terraform)

		terratest := new(config.TerratestConfig)
		operations.LoadObjectFromMap(config.TerratestConfigurationFileKey, configMap[0], terratest)

		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion
		testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

		a.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
			defer cleanup.Cleanup(a.T(), a.terraformOptions, keyPath)

			clusterIDs := provisioning.Provision(a.T(), a.client, a.rancherConfig, terraform, terratest, testUser, testPassword, clusterName, poolName, a.terraformOptions, configMap)
			provisioning.VerifyClustersState(a.T(), a.client, clusterIDs)

			provisioning.KubernetesUpgrade(a.T(), a.client, a.rancherConfig, terraform, terratest, testUser, testPassword, clusterName, poolName, a.terraformOptions, configMap)
			provisioning.VerifyClustersState(a.T(), a.client, clusterIDs)
		})
	}

	if a.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpAirgapProvisioningTestSuite(t *testing.T) {
	suite.Run(t, new(TfpAirgapProvisioningTestSuite))
}
