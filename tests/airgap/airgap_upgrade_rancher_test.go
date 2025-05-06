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
	"github.com/rancher/tfp-automation/framework/set/resources/upgrade"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TfpAirgapUpgradeRancherTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	session                    *session.Session
	cattleConfig               map[string]any
	rancherConfig              *rancher.Config
	terraformConfig            *config.TerraformConfig
	terratestConfig            *config.TerratestConfig
	standaloneTerraformOptions *terraform.Options
	upgradeTerraformOptions    *terraform.Options
	terraformOptions           *terraform.Options
	registry                   string
	bastion                    string
}

func (a *TfpAirgapUpgradeRancherTestSuite) TearDownSuite() {
	_, keyPath := rancher2.SetKeyPath(keypath.AirgapKeyPath, a.terraformConfig.Provider)
	cleanup.Cleanup(a.T(), a.standaloneTerraformOptions, keyPath)

	_, keyPath = rancher2.SetKeyPath(keypath.UpgradeKeyPath, a.terraformConfig.Provider)
	cleanup.Cleanup(a.T(), a.upgradeTerraformOptions, keyPath)
}

func (a *TfpAirgapUpgradeRancherTestSuite) SetupSuite() {
	a.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	a.rancherConfig, a.terraformConfig, a.terratestConfig = config.LoadTFPConfigs(a.cattleConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.AirgapKeyPath, a.terraformConfig.Provider)
	standaloneTerraformOptions := framework.Setup(a.T(), a.terraformConfig, a.terratestConfig, keyPath)
	a.standaloneTerraformOptions = standaloneTerraformOptions

	registry, bastion, err := airgap.CreateMainTF(a.T(), a.standaloneTerraformOptions, keyPath, a.terraformConfig, a.terratestConfig)
	require.NoError(a.T(), err)

	a.registry = registry
	a.bastion = bastion

	_, keyPath = rancher2.SetKeyPath(keypath.UpgradeKeyPath, a.terraformConfig.Provider)
	upgradeTerraformOptions := framework.Setup(a.T(), a.terraformConfig, a.terratestConfig, keyPath)

	a.upgradeTerraformOptions = upgradeTerraformOptions
}

func (a *TfpAirgapUpgradeRancherTestSuite) TfpSetupSuite() map[string]any {
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

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, "")
	terraformOptions := framework.Setup(a.T(), a.terraformConfig, a.terratestConfig, keyPath)
	a.terraformOptions = terraformOptions

	return a.cattleConfig
}

func (a *TfpAirgapUpgradeRancherTestSuite) TestTfpUpgradeAirgapRancher() {
	var clusterIDs []string

	a.provisionAndVerifyCluster("Pre-Upgrade Airgap ", clusterIDs, false)

	a.terraformConfig.Standalone.UpgradeAirgapRancher = true

	_, keyPath := rancher2.SetKeyPath(keypath.UpgradeKeyPath, a.terraformConfig.Provider)
	err := upgrade.CreateMainTF(a.T(), a.upgradeTerraformOptions, keyPath, a.terraformConfig, a.terratestConfig, "", "", a.bastion, a.registry)
	require.NoError(a.T(), err)

	provisioning.VerifyClustersState(a.T(), a.client, clusterIDs)

	a.provisionAndVerifyCluster("Post-Upgrade Airgap ", clusterIDs, true)

	if a.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func (a *TfpAirgapUpgradeRancherTestSuite) provisionAndVerifyCluster(name string, clusterIDs []string, deleteClusters bool) []string {
	tests := []struct {
		name   string
		module string
	}{
		{"RKE1", modules.AirgapRKE1},
		{"RKE2", modules.AirgapRKE2},
		{"RKE2 Windows", modules.AirgapRKE2Windows},
		{"K3S", modules.AirgapK3S},
	}

	newFile, rootBody, file := rancher2.InitializeMainTF()
	defer file.Close()

	customClusterNames := []string{}
	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		cattleConfig := a.TfpSetupSuite()
		configMap := []map[string]any{cattleConfig}

		_, err := operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(a.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, a.registry, configMap[0])
		require.NoError(a.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, a.registry, configMap[0])
		require.NoError(a.T(), err)

		provisioning.GetK8sVersion(a.T(), a.client, a.terratestConfig, a.terraformConfig, configs.DefaultK8sVersion, configMap)

		rancher, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		tt.name = name + tt.name + " Kubernetes version: " + terratest.KubernetesVersion

		a.Run((tt.name), func() {
			clusterIDs, customClusterNames = provisioning.Provision(a.T(), a.client, rancher, terraform, testUser, testPassword, a.terraformOptions, configMap, newFile, rootBody, file, false, true, true, customClusterNames)
			provisioning.VerifyClustersState(a.T(), a.client, clusterIDs)
			provisioning.VerifyRegistry(a.T(), a.client, clusterIDs[0], terraform)

			if strings.Contains(terraform.Module, modules.AirgapRKE2Windows) {
				clusterIDs, _ = provisioning.Provision(a.T(), a.client, rancher, terraform, testUser, testPassword, a.terraformOptions, configMap, newFile, rootBody, file, true, true, true, customClusterNames)
				provisioning.VerifyClustersState(a.T(), a.client, clusterIDs)
				provisioning.VerifyRegistry(a.T(), a.client, clusterIDs[0], terraform)
			}
		})
	}

	if deleteClusters {
		_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, "")
		cleanup.Cleanup(a.T(), a.terraformOptions, keyPath)
	}

	return clusterIDs
}

func TestTfpAirgapUpgradeRancherTestSuite(t *testing.T) {
	suite.Run(t, new(TfpAirgapUpgradeRancherTestSuite))
}
