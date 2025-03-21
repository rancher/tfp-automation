package sanity

import (
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/tests/actions/pipeline"
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
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TfpSanityTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	session                    *session.Session
	cattleConfig               map[string]any
	rancherConfig              *rancher.Config
	terraformConfig            *config.TerraformConfig
	terratestConfig            *config.TerratestConfig
	standaloneTerraformOptions *terraform.Options
	terraformOptions           *terraform.Options
}

func (t *TfpSanityTestSuite) TearDownSuite() {
	keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath)
	cleanup.Cleanup(t.T(), t.standaloneTerraformOptions, keyPath)
}

func (t *TfpSanityTestSuite) SetupSuite() {
	t.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	t.rancherConfig, t.terraformConfig, t.terratestConfig = config.LoadTFPConfigs(t.cattleConfig)

	keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath)
	standaloneTerraformOptions := framework.Setup(t.T(), t.terraformConfig, t.terratestConfig, keyPath)
	t.standaloneTerraformOptions = standaloneTerraformOptions

	err := resources.CreateMainTF(t.T(), t.standaloneTerraformOptions, keyPath, t.terraformConfig, t.terratestConfig)
	require.NoError(t.T(), err)
}

func (t *TfpSanityTestSuite) TfpSetupSuite() map[string]any {
	testSession := session.NewSession()
	t.session = testSession

	t.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	configMap, err := provisioning.UniquifyTerraform([]map[string]any{t.cattleConfig})
	require.NoError(t.T(), err)

	t.cattleConfig = configMap[0]
	t.rancherConfig, t.terraformConfig, t.terratestConfig = config.LoadTFPConfigs(t.cattleConfig)

	adminUser := &management.User{
		Username: "admin",
		Password: t.rancherConfig.AdminPassword,
	}

	userToken, err := token.GenerateUserToken(adminUser, t.rancherConfig.Host)
	require.NoError(t.T(), err)

	t.rancherConfig.AdminToken = userToken.Token

	client, err := rancher.NewClient(t.rancherConfig.AdminToken, testSession)
	require.NoError(t.T(), err)

	t.client = client
	t.client.RancherConfig.AdminToken = t.rancherConfig.AdminToken
	t.client.RancherConfig.AdminPassword = t.rancherConfig.AdminPassword
	t.client.RancherConfig.Host = t.rancherConfig.Host

	operations.ReplaceValue([]string{"rancher", "adminToken"}, t.rancherConfig.AdminToken, configMap[0])
	operations.ReplaceValue([]string{"rancher", "adminPassword"}, t.rancherConfig.AdminPassword, configMap[0])
	operations.ReplaceValue([]string{"rancher", "host"}, t.rancherConfig.Host, configMap[0])

	err = pipeline.PostRancherInstall(t.client, t.client.RancherConfig.AdminPassword)
	require.NoError(t.T(), err)

	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
	terraformOptions := framework.Setup(t.T(), t.terraformConfig, t.terratestConfig, keyPath)
	t.terraformOptions = terraformOptions

	return t.cattleConfig
}

func (t *TfpSanityTestSuite) TestTfpProvisioningSanity() {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		module    string
	}{
		{"Sanity RKE1", nodeRolesDedicated, "ec2_rke1"},
		{"Sanity RKE2", nodeRolesDedicated, "ec2_rke2"},
		{"Sanity RKE2 Windows", nil, "ec2_rke2_windows_custom"},
		{"Sanity K3S", nodeRolesDedicated, "ec2_k3s"},
	}

	for _, tt := range tests {
		cattleConfig := t.TfpSetupSuite()
		configMap := []map[string]any{cattleConfig}

		operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])

		provisioning.GetK8sVersion(t.T(), t.client, t.terratestConfig, t.terraformConfig, configs.DefaultK8sVersion, configMap)

		_, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion
		testUser, testPassword := configs.CreateTestCredentials()

		t.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
			defer cleanup.Cleanup(t.T(), t.terraformOptions, keyPath)

			clusterIDs := provisioning.Provision(t.T(), t.client, t.rancherConfig, terraform, terratest, testUser, testPassword, t.terraformOptions, configMap, false)
			provisioning.VerifyClustersState(t.T(), t.client, clusterIDs)

			if strings.Contains(terraform.Module, modules.CustomEC2RKE2Windows) {
				clusterIDs := provisioning.Provision(t.T(), t.client, t.rancherConfig, terraform, terratest, testUser, testPassword, t.terraformOptions, configMap, true)
				provisioning.VerifyClustersState(t.T(), t.client, clusterIDs)
			}
		})
	}

	if t.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpSanityTestSuite(t *testing.T) {
	suite.Run(t, new(TfpSanityTestSuite))
}
