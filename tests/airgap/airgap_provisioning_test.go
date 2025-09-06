package airgap

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/infrastructure"
	"github.com/sirupsen/logrus"
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
	standaloneConfig           *config.Standalone
	standaloneTerraformOptions *terraform.Options
	terraformOptions           *terraform.Options
	registry                   string
}

func (a *TfpAirgapProvisioningTestSuite) TearDownSuite() {
	_, keyPath := rancher2.SetKeyPath(keypath.AirgapKeyPath, a.terratestConfig.PathToRepo, a.terraformConfig.Provider)
	cleanup.Cleanup(a.T(), a.standaloneTerraformOptions, keyPath)
}

func (a *TfpAirgapProvisioningTestSuite) SetupSuite() {
	testSession := session.NewSession()
	a.session = testSession

	a.client, a.registry, _, a.standaloneTerraformOptions, a.terraformOptions, a.cattleConfig = infrastructure.SetupAirgapRancher(a.T(), a.session, keypath.AirgapKeyPath)
	a.rancherConfig, a.terraformConfig, a.terratestConfig, a.standaloneConfig = config.LoadTFPConfigs(a.cattleConfig)

	provisioning.VerifyRancherVersion(a.T(), a.rancherConfig.Host, a.standaloneConfig.RancherTagVersion)
}

func (a *TfpAirgapProvisioningTestSuite) TestTfpAirgapProvisioning() {
	tests := []struct {
		name   string
		module string
	}{
		{"Airgap_RKE2", modules.AirgapRKE2},
		{"Airgap_RKE2_Windows_2019", modules.AirgapRKE2Windows2019},
		{"Airgap_RKE2_Windows_2022", modules.AirgapRKE2Windows2022},
		{"Airgap_K3S", modules.AirgapK3S},
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

		rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

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

		params := tfpQase.GetProvisioningSchemaParams(a.client, configMap[0])
		err = qase.UpdateSchemaParameters(tt.name, params)
		if err != nil {
			logrus.Warningf("Failed to upload schema parameters %s", err)
		}
	}

	if a.terratestConfig.LocalQaseReporting {
		results.ReportTest(a.terratestConfig)
	}
}

func TestTfpAirgapProvisioningTestSuite(t *testing.T) {
	suite.Run(t, new(TfpAirgapProvisioningTestSuite))
}
