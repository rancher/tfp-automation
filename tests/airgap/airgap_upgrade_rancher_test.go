package airgap

import (
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	steveV1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	clusterActions "github.com/rancher/tests/actions/clusters"
	provisioningActions "github.com/rancher/tests/actions/provisioning"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	nested "github.com/rancher/tfp-automation/tests/extensions/nestedModules"
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

	if a.tunnel != nil {
		a.tunnel.StopBastionSSHTunnel()
	}
}

func (a *TfpAirgapUpgradeRancherTestSuite) SetupSuite() {
	testSession := session.NewSession()
	a.session = testSession

	a.client, a.registry, a.bastion, a.standaloneTerraformOptions, a.terraformOptions, a.cattleConfig, a.tunnel = ranchers.SetupAirgapRancher(a.T(), a.session, keypath.AirgapKeyPath)
	a.rancherConfig, a.terraformConfig, a.terratestConfig, a.standaloneConfig = config.LoadTFPConfigs(a.cattleConfig)
}

func (a *TfpAirgapUpgradeRancherTestSuite) TestTfpUpgradeAirgapRancher() {
	standardUserClient, standardToken, testUser, testPassword := ranchers.SetupResources(a.T(), a.client, a.rancherConfig, a.terratestConfig, a.terraformOptions)

	a.rancherConfig, a.terraformConfig, a.terratestConfig, _ = config.LoadTFPConfigs(a.cattleConfig)
	nestedRancherModuleDir := a.provisionAndVerifyCluster("Airgap_Pre_Rancher_Upgrade", standardUserClient, standardToken, testUser, testPassword)

	a.client, a.cattleConfig, a.terraformOptions, a.upgradeTerraformOptions = ranchers.UpgradeAirgapRancher(a.T(), a.client, a.bastion, a.registry, a.session, a.cattleConfig, a.tunnel)

	ranchers.CleanupDownstreamClusters(a.T(), a.client, a.terraformConfig)
	os.RemoveAll(nestedRancherModuleDir)

	standardUserClient, standardToken, testUser, testPassword = ranchers.SetupResources(a.T(), a.client, a.rancherConfig, a.terratestConfig, a.terraformOptions)

	a.rancherConfig, a.terraformConfig, a.terratestConfig, _ = config.LoadTFPConfigs(a.cattleConfig)
	nestedRancherModuleDir = a.provisionAndVerifyCluster("Airgap_Post_Rancher_Upgrade", standardUserClient, standardToken, testUser, testPassword)

	ranchers.CleanupDownstreamClusters(a.T(), a.client, a.terraformConfig)
	os.RemoveAll(nestedRancherModuleDir)

	if a.terratestConfig.LocalQaseReporting {
		results.ReportTest(a.terratestConfig)
	}
}

func (a *TfpAirgapUpgradeRancherTestSuite) provisionAndVerifyCluster(name string, standardUserClient *rancher.Client, standardToken,
	testUser, testPassword string) string {
	var clusterIDs []string
	var nestedRancherModuleDir string
	var clusters []*steveV1.SteveAPIObject

	customClusterNames := []string{}

	tests := []struct {
		name   string
		module string
	}{
		{name + "_RKE2", modules.AirgapRKE2},
		{name + "_RKE2_Windows_2019", modules.AirgapRKE2Windows2019},
		{name + "_RKE2_Windows_2022", modules.AirgapRKE2Windows2022},
		{name + "_K3S", modules.AirgapK3S},
	}

	a.T().Run(name, func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(a.terraformConfig, a.terratestConfig, a.terraformOptions, tt.name, configs.NestedRancherModuleDir)
				require.NoError(t, err)

				newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
				defer file.Close()

				cattleConfig, err := provisioning.UniquifyTerraform(a.cattleConfig)
				require.NoError(t, err)

				_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, cattleConfig)
				require.NoError(t, err)

				_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, cattleConfig)
				require.NoError(t, err)

				_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, a.registry, cattleConfig)
				require.NoError(t, err)

				_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, a.registry, cattleConfig)
				require.NoError(t, err)

				provisioning.GetK8sVersion(a.client, cattleConfig)

				rancher, terraform, terratest, _ := config.LoadTFPConfigs(cattleConfig)

				clusters, customClusterNames = provisioning.Provision(t, a.client, standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, newFile, rootBody, file, false, true, true, clusterIDs, customClusterNames, nestedRancherModuleDir)
				err = provisioningActions.VerifyClusterReady(a.client, clusters[0])
				require.NoError(t, err)

				err = clusterActions.VerifyServiceAccountTokenSecret(a.client, clusters[0].Name)
				require.NoError(t, err)
				cluster, err := a.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraform.ResourcePrefix)
				require.NoError(t, err)

				err = pods.VerifyClusterPods(a.client, cluster)
				require.NoError(t, err)

				if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
					clusters, customClusterNames = provisioning.Provision(t, a.client, standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, newFile, rootBody, file, true, true, true, clusterIDs, customClusterNames, nestedRancherModuleDir)
					err = provisioningActions.VerifyClusterReady(a.client, clusters[0])
					require.NoError(t, err)

					err = clusterActions.VerifyServiceAccountTokenSecret(a.client, clusters[0].Name)
					require.NoError(t, err)
					err = pods.VerifyClusterPods(a.client, cluster)
					require.NoError(t, err)
				}

				params := tfpQase.GetProvisioningSchemaParams(cattleConfig)
				err = qase.UpdateSchemaParameters(tt.name, params)
				if err != nil {
					logrus.Warningf("Failed to upload schema parameters %s", err)
				}
			})
		}
	})

	return nestedRancherModuleDir
}

func TestTfpAirgapUpgradeRancherTestSuite(t *testing.T) {
	suite.Run(t, new(TfpAirgapUpgradeRancherTestSuite))
}
