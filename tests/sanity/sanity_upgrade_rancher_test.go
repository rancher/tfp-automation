package sanity

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

type TfpSanityUpgradeRancherTestSuite struct {
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
	serverNodeOne              string
}

func (s *TfpSanityUpgradeRancherTestSuite) TearDownSuite() {
	_, keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath, s.terratestConfig.PathToRepo, s.terraformConfig.Provider)
	cleanup.Cleanup(s.T(), s.standaloneTerraformOptions, keyPath)

	_, keyPath = rancher2.SetKeyPath(keypath.UpgradeKeyPath, s.terratestConfig.PathToRepo, s.terraformConfig.Provider)
	cleanup.Cleanup(s.T(), s.upgradeTerraformOptions, keyPath)
}

func (s *TfpSanityUpgradeRancherTestSuite) SetupSuite() {
	testSession := session.NewSession()
	s.session = testSession

	s.client, s.serverNodeOne, s.standaloneTerraformOptions, s.terraformOptions, s.cattleConfig = infrastructure.SetupRancher(s.T(), s.session, keypath.SanityKeyPath)
}

func (s *TfpSanityUpgradeRancherTestSuite) TestTfpUpgradeRancher() {
	var clusterIDs []string

	testUser, testPassword := configs.CreateTestCredentials()

	s.rancherConfig, s.terraformConfig, s.terratestConfig = config.LoadTFPConfigs(s.cattleConfig)
	s.provisionAndVerifyCluster("Pre-Upgrade Sanity ", clusterIDs, false, testUser, testPassword)

	s.client, s.cattleConfig, s.terraformOptions, s.upgradeTerraformOptions = infrastructure.UpgradeRancher(s.T(), s.client, s.serverNodeOne, s.session, s.cattleConfig)

	s.rancherConfig, s.terraformConfig, s.terratestConfig = config.LoadTFPConfigs(s.cattleConfig)
	s.provisionAndVerifyCluster("Post-Upgrade Sanity ", clusterIDs, true, testUser, testPassword)

	if s.terratestConfig.LocalQaseReporting {
		qase.ReportTest(s.terratestConfig)
	}
}

func (s *TfpSanityUpgradeRancherTestSuite) provisionAndVerifyCluster(name string, clusterIDs []string, deleteClusters bool,
	testUser, testPassword string) []string {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		module    string
	}{
		{"RKE2", nodeRolesDedicated, modules.EC2RKE2},
		{"RKE2 Windows 2019", nil, modules.CustomEC2RKE2Windows2019},
		{"RKE2 Windows 2022", nil, modules.CustomEC2RKE2Windows2022},
		{"K3S", nodeRolesDedicated, modules.EC2K3s},
	}

	newFile, rootBody, file := rancher2.InitializeMainTF(s.terratestConfig)
	defer file.Close()

	customClusterNames := []string{}

	for _, tt := range tests {
		configMap, err := provisioning.UniquifyTerraform([]map[string]any{s.cattleConfig})
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, s.client.RancherConfig.AdminToken, configMap[0])
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(s.T(), err)

		provisioning.GetK8sVersion(s.T(), s.client, s.terratestConfig, s.terraformConfig, configs.SecondHighestVersion, configMap)

		rancher, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		currentDate := time.Now().Format("2006-01-02 03:04PM")
		tt.name = name + tt.name + " Kubernetes version: " + terratest.KubernetesVersion + " " + currentDate

		s.Run((tt.name), func() {
			clusterIDs, customClusterNames = provisioning.Provision(s.T(), s.client, rancher, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file, false, true, true, customClusterNames)
			provisioning.VerifyClustersState(s.T(), s.client, clusterIDs)

			if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
				clusterIDs, customClusterNames = provisioning.Provision(s.T(), s.client, rancher, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file, true, true, true, customClusterNames)
				provisioning.VerifyClustersState(s.T(), s.client, clusterIDs)
			}

			if deleteClusters {
				provisioning.KubernetesUpgrade(s.T(), s.client, rancher, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file, false)
				provisioning.VerifyClustersState(s.T(), s.client, clusterIDs)
			}
		})
	}

	if deleteClusters {
		_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, s.terratestConfig.PathToRepo, "")
		cleanup.Cleanup(s.T(), s.terraformOptions, keyPath)
	}

	return clusterIDs
}

func TestTfpSanityUpgradeRancherTestSuite(t *testing.T) {
	suite.Run(t, new(TfpSanityUpgradeRancherTestSuite))
}
