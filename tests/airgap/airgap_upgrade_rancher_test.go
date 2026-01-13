package airgap

import (
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/extensions/ssh"
	"github.com/rancher/tfp-automation/tests/infrastructure/ranchers"
	"github.com/sirupsen/logrus"
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
	standaloneConfig           *config.Standalone
	standaloneTerraformOptions *terraform.Options
	upgradeTerraformOptions    *terraform.Options
	terraformOptions           *terraform.Options
	registry                   string
	bastion                    string
	tunnel                     *ssh.BastionSSHTunnel
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

	a.client, a.registry, a.bastion, a.standaloneTerraformOptions, a.terraformOptions, a.cattleConfig, a.tunnel = ranchers.SetupAirgapRancher(a.T(), a.session, keypath.AirgapKeyPath)
	a.rancherConfig, a.terraformConfig, a.terratestConfig, a.standaloneConfig = config.LoadTFPConfigs(a.cattleConfig)
}

func (a *TfpAirgapUpgradeRancherTestSuite) TestTfpUpgradeAirgapRancher() {
	var clusterIDs []string

	standardUserClient, newFile, rootBody, file, standardToken, testUser, testPassword := ranchers.SetupResources(a.T(), a.client, a.rancherConfig, a.terratestConfig, a.terraformOptions)

	a.rancherConfig, a.terraformConfig, a.terratestConfig, _ = config.LoadTFPConfigs(a.cattleConfig)
	allClusterIDs := a.provisionAndVerifyCluster("Airgap_Pre_Rancher_Upgrade_", clusterIDs, newFile, rootBody, file, standardUserClient, standardToken, testUser, testPassword)

	a.client, a.cattleConfig, a.terraformOptions, a.upgradeTerraformOptions = ranchers.UpgradeAirgapRancher(a.T(), a.client, a.bastion, a.registry, a.session, a.cattleConfig, a.tunnel)
	provisioning.VerifyClustersState(a.T(), a.client, allClusterIDs)

	ranchers.CleanupPreUpgradeClusters(a.T(), a.client, allClusterIDs, a.terraformConfig)

	standardUserClient, newFile, rootBody, file, standardToken, testUser, testPassword = ranchers.SetupResources(a.T(), a.client, a.rancherConfig, a.terratestConfig, a.terraformOptions)

	a.rancherConfig, a.terraformConfig, a.terratestConfig, _ = config.LoadTFPConfigs(a.cattleConfig)
	a.provisionAndVerifyCluster("Airgap_Post_Rancher_Upgrade_", nil, newFile, rootBody, file, standardUserClient, standardToken, testUser, testPassword)

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, a.terratestConfig.PathToRepo, "")
	cleanup.Cleanup(a.T(), a.terraformOptions, keyPath)

	if a.terratestConfig.LocalQaseReporting {
		results.ReportTest(a.terratestConfig)
	}
}

func (a *TfpAirgapUpgradeRancherTestSuite) provisionAndVerifyCluster(name string, clusterIDs []string, newFile *hclwrite.File, rootBody *hclwrite.Body,
	file *os.File, standardUserClient *rancher.Client, standardToken, testUser, testPassword string) []string {
	customClusterNames := []string{}

	tests := []struct {
		name   string
		module string
	}{
		{"RKE2", modules.AirgapRKE2},
		{"RKE2_Windows_2019", modules.AirgapRKE2Windows2019},
		{"RKE2_Windows_2022", modules.AirgapRKE2Windows2022},
		{"K3S", modules.AirgapK3S},
	}

	for _, tt := range tests {
		configMap, err := provisioning.UniquifyTerraform([]map[string]any{a.cattleConfig})
		require.NoError(a.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, configMap[0])
		require.NoError(a.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(a.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, a.registry, configMap[0])
		require.NoError(a.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, a.registry, configMap[0])
		require.NoError(a.T(), err)

		provisioning.GetK8sVersion(a.T(), standardUserClient, a.terratestConfig, a.terraformConfig, configMap)

		rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

		tt.name = name + tt.name

		a.Run((tt.name), func() {
			clusterIDs, customClusterNames = provisioning.Provision(a.T(), a.client, standardUserClient, rancher, terraform, terratest, testUser, testPassword, a.terraformOptions, configMap, newFile, rootBody, file, false, true, true, clusterIDs, customClusterNames)
			provisioning.VerifyClustersState(a.T(), a.client, clusterIDs)
			provisioning.VerifyServiceAccountTokenSecret(a.T(), a.client, clusterIDs)

			cluster, err := a.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraform.ResourcePrefix)
			require.NoError(a.T(), err)

			err = pods.VerifyClusterPods(a.client, cluster)
			require.NoError(a.T(), err)

			if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
				clusterIDs, customClusterNames = provisioning.Provision(a.T(), a.client, standardUserClient, rancher, terraform, terratest, testUser, testPassword, a.terraformOptions, configMap, newFile, rootBody, file, true, true, true, clusterIDs, customClusterNames)
				provisioning.VerifyClustersState(a.T(), a.client, clusterIDs)
				provisioning.VerifyServiceAccountTokenSecret(a.T(), a.client, clusterIDs)

				err = pods.VerifyClusterPods(a.client, cluster)
				require.NoError(a.T(), err)
			}
		})

		params := tfpQase.GetProvisioningSchemaParams(configMap[0])
		err = qase.UpdateSchemaParameters(tt.name, params)
		if err != nil {
			logrus.Warningf("Failed to upload schema parameters %s", err)
		}
	}

	return ranchers.UniqueStrings(clusterIDs)
}

func TestTfpAirgapUpgradeRancherTestSuite(t *testing.T) {
	suite.Run(t, new(TfpAirgapUpgradeRancherTestSuite))
}
