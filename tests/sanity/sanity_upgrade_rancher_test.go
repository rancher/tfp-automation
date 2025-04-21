package sanity

import (
	"os"
	"strings"
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
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	resources "github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/rancher/tfp-automation/framework/set/resources/upgrade"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
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
	keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath, s.terraformConfig.Provider)
	cleanup.Cleanup(s.T(), s.standaloneTerraformOptions, keyPath)

	keyPath = rancher2.SetKeyPath(keypath.UpgradeKeyPath, s.terraformConfig.Provider)
	cleanup.Cleanup(s.T(), s.upgradeTerraformOptions, keyPath)
}

func (s *TfpSanityUpgradeRancherTestSuite) SetupSuite() {
	s.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	s.rancherConfig, s.terraformConfig, s.terratestConfig = config.LoadTFPConfigs(s.cattleConfig)

	keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath, s.terraformConfig.Provider)
	standaloneTerraformOptions := framework.Setup(s.T(), s.terraformConfig, s.terratestConfig, keyPath)
	s.standaloneTerraformOptions = standaloneTerraformOptions

	serverNodeOne, err := resources.CreateMainTF(s.T(), s.standaloneTerraformOptions, keyPath, s.terraformConfig, s.terratestConfig)
	require.NoError(s.T(), err)

	s.serverNodeOne = serverNodeOne

	keyPath = rancher2.SetKeyPath(keypath.UpgradeKeyPath, s.terraformConfig.Provider)
	upgradeTerraformOptions := framework.Setup(s.T(), s.terraformConfig, s.terratestConfig, keyPath)

	s.upgradeTerraformOptions = upgradeTerraformOptions
}

func (s *TfpSanityUpgradeRancherTestSuite) TfpSetupSuite() map[string]any {
	testSession := session.NewSession()
	s.session = testSession

	s.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	configMap, err := provisioning.UniquifyTerraform([]map[string]any{s.cattleConfig})
	require.NoError(s.T(), err)

	s.cattleConfig = configMap[0]
	s.rancherConfig, s.terraformConfig, s.terratestConfig = config.LoadTFPConfigs(s.cattleConfig)

	adminUser := &management.User{
		Username: "admin",
		Password: s.rancherConfig.AdminPassword,
	}

	userToken, err := token.GenerateUserToken(adminUser, s.rancherConfig.Host)
	require.NoError(s.T(), err)

	s.rancherConfig.AdminToken = userToken.Token

	client, err := rancher.NewClient(s.rancherConfig.AdminToken, testSession)
	require.NoError(s.T(), err)

	s.client = client
	s.client.RancherConfig.AdminToken = s.rancherConfig.AdminToken
	s.client.RancherConfig.AdminPassword = s.rancherConfig.AdminPassword
	s.client.RancherConfig.Host = s.rancherConfig.Host

	operations.ReplaceValue([]string{"rancher", "adminToken"}, s.rancherConfig.AdminToken, configMap[0])
	operations.ReplaceValue([]string{"rancher", "adminPassword"}, s.rancherConfig.AdminPassword, configMap[0])
	operations.ReplaceValue([]string{"rancher", "host"}, s.rancherConfig.Host, configMap[0])

	err = pipeline.PostRancherInstall(s.client, s.client.RancherConfig.AdminPassword)
	require.NoError(s.T(), err)

	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, "")
	terraformOptions := framework.Setup(s.T(), s.terraformConfig, s.terratestConfig, keyPath)
	s.terraformOptions = terraformOptions

	return s.cattleConfig
}

func (s *TfpSanityUpgradeRancherTestSuite) TestTfpUpgradeRancher() {
	var clusterIDs []string

	s.provisionAndVerifyCluster("Pre-Upgrade Sanity ", clusterIDs, false)

	s.terraformConfig.Standalone.UpgradeRancher = true

	keyPath := rancher2.SetKeyPath(keypath.UpgradeKeyPath, s.terraformConfig.Provider)
	err := upgrade.CreateMainTF(s.T(), s.upgradeTerraformOptions, keyPath, s.terraformConfig, s.terratestConfig, s.serverNodeOne, "", "", "")
	require.NoError(s.T(), err)

	provisioning.VerifyClustersState(s.T(), s.client, clusterIDs)

	s.provisionAndVerifyCluster("Post-Upgrade Sanity ", clusterIDs, true)

	if s.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func (s *TfpSanityUpgradeRancherTestSuite) provisionAndVerifyCluster(name string, clusterIDs []string, deleteClusters bool) []string {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		module    string
	}{
		{"RKE1", nodeRolesDedicated, modules.EC2RKE1},
		{"RKE2", nodeRolesDedicated, modules.EC2RKE2},
		{"RKE2 Windows", nil, modules.CustomEC2RKE2Windows},
		{"K3S", nodeRolesDedicated, modules.EC2K3s},
	}

	newFile, rootBody, file := rancher2.InitializeMainTF()
	defer file.Close()

	customClusterNames := []string{}
	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		cattleConfig := s.TfpSetupSuite()
		configMap := []map[string]any{cattleConfig}

		_, err := operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(s.T(), err)

		provisioning.GetK8sVersion(s.T(), s.client, s.terratestConfig, s.terraformConfig, configs.DefaultK8sVersion, configMap)

		rancher, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		tt.name = name + tt.name + " Kubernetes version: " + terratest.KubernetesVersion

		s.Run((tt.name), func() {
			clusterIDs, customClusterNames = provisioning.Provision(s.T(), s.client, rancher, terraform, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file, false, true, true, customClusterNames)
			provisioning.VerifyClustersState(s.T(), s.client, clusterIDs)

			if strings.Contains(terraform.Module, modules.CustomEC2RKE2Windows) {
				clusterIDs, _ = provisioning.Provision(s.T(), s.client, rancher, terraform, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file, true, true, true, customClusterNames)
				provisioning.VerifyClustersState(s.T(), s.client, clusterIDs)
			}
		})
	}

	if deleteClusters {
		keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, "")
		cleanup.Cleanup(s.T(), s.terraformOptions, keyPath)
	}

	return clusterIDs
}

func TestTfpSanityUpgradeRancherTestSuite(t *testing.T) {
	suite.Run(t, new(TfpSanityUpgradeRancherTestSuite))
}
