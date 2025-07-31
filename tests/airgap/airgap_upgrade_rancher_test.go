package airgap

import (
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/infrastructure"
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
	_, keyPath := rancher2.SetKeyPath(keypath.AirgapKeyPath, a.terratestConfig.PathToRepo, a.terraformConfig.Provider)
	cleanup.Cleanup(a.T(), a.standaloneTerraformOptions, keyPath)

	_, keyPath = rancher2.SetKeyPath(keypath.UpgradeKeyPath, a.terratestConfig.PathToRepo, a.terraformConfig.Provider)
	cleanup.Cleanup(a.T(), a.upgradeTerraformOptions, keyPath)
}

func (a *TfpAirgapUpgradeRancherTestSuite) SetupSuite() {
	testSession := session.NewSession()
	a.session = testSession

	a.client, a.registry, a.bastion, a.standaloneTerraformOptions, a.terraformOptions, a.cattleConfig = infrastructure.SetupAirgapRancher(a.T(), a.session, keypath.AirgapKeyPath)
}

func (a *TfpAirgapUpgradeRancherTestSuite) TestTfpUpgradeAirgapRancher() {
	var clusterIDs []string

	testUser, testPassword := configs.CreateTestCredentials()

	a.rancherConfig, a.terraformConfig, a.terratestConfig = config.LoadTFPConfigs(a.cattleConfig)
	a.provisionAndVerifyCluster("Pre-Upgrade Airgap ", clusterIDs, false, testUser, testPassword)

	a.client, a.cattleConfig, a.terraformOptions, a.upgradeTerraformOptions = infrastructure.UpgradeAirgapRancher(a.T(), a.client, a.bastion, a.registry, a.session, a.cattleConfig)

	a.rancherConfig, a.terraformConfig, a.terratestConfig = config.LoadTFPConfigs(a.cattleConfig)
	a.provisionAndVerifyCluster("Post-Upgrade Airgap ", clusterIDs, true, testUser, testPassword)

	if a.terratestConfig.LocalQaseReporting {
		qase.ReportTest(a.terratestConfig)
	}
}

func (a *TfpAirgapUpgradeRancherTestSuite) provisionAndVerifyCluster(name string, clusterIDs []string, deleteClusters bool,
	testUser, testPassword string) []string {
	tests := []struct {
		name   string
		module string
	}{
		{"RKE2", modules.AirgapRKE2},
		{"RKE2 Windows 2019", modules.AirgapRKE2Windows2019},
		{"RKE2 Windows 2022", modules.AirgapRKE2Windows2022},
		{"K3S", modules.AirgapK3S},
	}

	newFile, rootBody, file := rancher2.InitializeMainTF(a.terratestConfig)
	defer file.Close()

	customClusterNames := []string{}

	for _, tt := range tests {
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
		tt.name = name + tt.name + " Kubernetes version: " + terratest.KubernetesVersion + " " + currentDate

		a.Run((tt.name), func() {
			clusterIDs, customClusterNames = provisioning.Provision(a.T(), a.client, rancher, terraform, terratest, testUser, testPassword, a.terraformOptions, configMap, newFile, rootBody, file, false, true, true, customClusterNames)
			provisioning.VerifyClustersState(a.T(), a.client, clusterIDs)

			if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
				clusterIDs, customClusterNames = provisioning.Provision(a.T(), a.client, rancher, terraform, terratest, testUser, testPassword, a.terraformOptions, configMap, newFile, rootBody, file, true, true, true, customClusterNames)
				provisioning.VerifyClustersState(a.T(), a.client, clusterIDs)
			}
		})
	}

	if deleteClusters {
		_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, a.terratestConfig.PathToRepo, "")
		cleanup.Cleanup(a.T(), a.terraformOptions, keyPath)
	}

	return clusterIDs
}

func TestTfpAirgapUpgradeRancherTestSuite(t *testing.T) {
	suite.Run(t, new(TfpAirgapUpgradeRancherTestSuite))
}
