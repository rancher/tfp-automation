package airgap

import (
	"os"
	"strings"
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
	"github.com/rancher/tfp-automation/defaults/modules"
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

	registry, _, err := airgap.CreateMainTF(a.T(), a.standaloneTerraformOptions, keyPath, a.terraformConfig, a.terratestConfig)
	require.NoError(a.T(), err)

	a.registry = registry
}

func (a *TfpAirgapProvisioningTestSuite) TfpSetupSuite() map[string]any {
	testSession := session.NewSession()
	a.session = testSession

	a.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	configMap, err := provisioning.UniquifyTerraform([]map[string]any{a.cattleConfig})
	require.NoError(a.T(), err)

	a.cattleConfig = configMap[0]
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

	operations.ReplaceValue([]string{"rancher", "adminToken"}, a.rancherConfig.AdminToken, configMap[0])
	operations.ReplaceValue([]string{"rancher", "adminPassword"}, a.rancherConfig.AdminPassword, configMap[0])
	operations.ReplaceValue([]string{"rancher", "host"}, a.rancherConfig.Host, configMap[0])

	err = pipeline.PostRancherInstall(a.client, a.client.RancherConfig.AdminPassword)
	require.NoError(a.T(), err)

	a.client.RancherConfig.Host = a.rancherConfig.Host

	operations.ReplaceValue([]string{"rancher", "host"}, a.rancherConfig.Host, configMap[0])

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
		{"Airgap RKE2 Windows", "airgap_rke2_windows"},
		{"Airgap K3S", "airgap_k3s"},
	}

	for _, tt := range tests {
		cattleConfig := a.TfpSetupSuite()
		configMap := []map[string]any{cattleConfig}

		operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, a.registry, configMap[0])
		operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, a.registry, configMap[0])

		provisioning.GetK8sVersion(a.T(), a.client, a.terratestConfig, a.terraformConfig, configs.DefaultK8sVersion, configMap)

		_, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion
		testUser, testPassword := configs.CreateTestCredentials()

		a.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
			defer cleanup.Cleanup(a.T(), a.terraformOptions, keyPath)

			clusterIDs := provisioning.Provision(a.T(), a.client, a.rancherConfig, terraform, terratest, testUser, testPassword, a.terraformOptions, configMap, false)
			provisioning.VerifyClustersState(a.T(), a.client, clusterIDs)
			provisioning.VerifyRegistry(a.T(), a.client, clusterIDs[0], terraform)

			if strings.Contains(terraform.Module, modules.AirgapRKE2Windows) {
				clusterIDs := provisioning.Provision(a.T(), a.client, a.rancherConfig, terraform, terratest, testUser, testPassword, a.terraformOptions, configMap, true)
				provisioning.VerifyClustersState(a.T(), a.client, clusterIDs)
				provisioning.VerifyRegistry(a.T(), a.client, clusterIDs[0], terraform)
			}
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
		{"Upgrading Airgap RKE2 Windows", "airgap_rke2_windows"},
		{"Upgrading Airgap K3S", "airgap_k3s"},
	}

	for _, tt := range tests {
		cattleConfig := a.TfpSetupSuite()
		configMap := []map[string]any{cattleConfig}

		operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, a.registry, configMap[0])
		operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, a.registry, configMap[0])

		provisioning.GetK8sVersion(a.T(), a.client, a.terratestConfig, a.terraformConfig, configs.SecondHighestVersion, configMap)

		_, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion
		testUser, testPassword := configs.CreateTestCredentials()

		a.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
			defer cleanup.Cleanup(a.T(), a.terraformOptions, keyPath)

			clusterIDs := provisioning.Provision(a.T(), a.client, a.rancherConfig, terraform, terratest, testUser, testPassword, a.terraformOptions, configMap, false)
			provisioning.VerifyClustersState(a.T(), a.client, clusterIDs)
			provisioning.VerifyRegistry(a.T(), a.client, clusterIDs[0], terraform)

			if strings.Contains(terraform.Module, modules.AirgapRKE2Windows) {
				clusterIDs := provisioning.Provision(a.T(), a.client, a.rancherConfig, terraform, terratest, testUser, testPassword, a.terraformOptions, configMap, true)
				provisioning.VerifyClustersState(a.T(), a.client, clusterIDs)
				provisioning.VerifyRegistry(a.T(), a.client, clusterIDs[0], terraform)
			}

			provisioning.KubernetesUpgrade(a.T(), a.client, a.rancherConfig, terraform, terratest, testUser, testPassword, a.terraformOptions, configMap)
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
