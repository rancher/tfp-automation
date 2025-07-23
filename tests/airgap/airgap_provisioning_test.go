package airgap

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/airgap"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/infrastructure"
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
	_, keyPath := rancher2.SetKeyPath(keypath.AirgapKeyPath, a.terratestConfig.PathToRepo, a.terraformConfig.Provider)
	cleanup.Cleanup(a.T(), a.standaloneTerraformOptions, keyPath)
}

func (a *TfpAirgapProvisioningTestSuite) SetupSuite() {
	a.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	a.rancherConfig, a.terraformConfig, a.terratestConfig = config.LoadTFPConfigs(a.cattleConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.AirgapKeyPath, a.terratestConfig.PathToRepo, a.terraformConfig.Provider)
	standaloneTerraformOptions := framework.Setup(a.T(), a.terraformConfig, a.terratestConfig, keyPath)
	a.standaloneTerraformOptions = standaloneTerraformOptions

	registry, _, err := airgap.CreateMainTF(a.T(), a.standaloneTerraformOptions, keyPath, a.rancherConfig, a.terraformConfig, a.terratestConfig)
	require.NoError(a.T(), err)

	a.registry = registry

	testSession := session.NewSession()
	a.session = testSession

	client, err := infrastructure.PostRancherSetup(a.T(), a.rancherConfig, testSession, a.terraformConfig.Standalone.AirgapInternalFQDN, false, true)
	if err != nil && *a.rancherConfig.Cleanup {
		cleanup.Cleanup(a.T(), a.standaloneTerraformOptions, keyPath)
	}

	a.client = client

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, a.terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(a.T(), a.terraformConfig, a.terratestConfig, keyPath)
	a.terraformOptions = terraformOptions
}

func (a *TfpAirgapProvisioningTestSuite) TestTfpAirgapProvisioning() {
	tests := []struct {
		name   string
		module string
	}{
		{"Airgap RKE2", modules.AirgapRKE2},
		{"Airgap RKE2 Windows 2019", modules.AirgapRKE2Windows2019},
		{"Airgap RKE2 Windows 2022", modules.AirgapRKE2Windows2022},
		{"Airgap K3S", modules.AirgapK3S},
	}

	customClusterNames := []string{}
	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(a.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{a.cattleConfig})
		require.NoError(a.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, a.client.RancherConfig.AdminToken, configMap[0])
		require.NoError(a.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(a.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, a.registry, configMap[0])
		require.NoError(a.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, a.registry, configMap[0])
		require.NoError(a.T(), err)

		provisioning.GetK8sVersion(a.T(), a.client, a.terratestConfig, a.terraformConfig, configs.DefaultK8sVersion, configMap)

		rancher, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		currentDate := time.Now().Format("2006-01-02 03:04PM")
		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion + " " + currentDate

		a.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, a.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(a.T(), a.terraformOptions, keyPath)

			clusterIDs, customClusterNames := provisioning.Provision(a.T(), a.client, rancher, terraform, terratest, testUser, testPassword, a.terraformOptions, configMap, newFile, rootBody, file, false, false, true, customClusterNames)
			provisioning.VerifyClustersState(a.T(), a.client, clusterIDs)

			if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
				clusterIDs, _ = provisioning.Provision(a.T(), a.client, rancher, terraform, terratest, testUser, testPassword, a.terraformOptions, configMap, newFile, rootBody, file, true, true, true, customClusterNames)
				provisioning.VerifyClustersState(a.T(), a.client, clusterIDs)
			}
		})
	}

	if a.terratestConfig.LocalQaseReporting {
		qase.ReportTest(a.terratestConfig)
	}
}

func TestTfpAirgapProvisioningTestSuite(t *testing.T) {
	suite.Run(t, new(TfpAirgapProvisioningTestSuite))
}
