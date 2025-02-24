package sanity

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/token"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
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
	adminUser                  *management.User
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

	resources.CreateMainTF(t.T(), t.standaloneTerraformOptions, keyPath, t.terraformConfig, t.terratestConfig)
}

func (t *TfpSanityTestSuite) TfpSetupSuite() map[string]any {
	testSession := session.NewSession()
	t.session = testSession

	t.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	t.rancherConfig, t.terraformConfig, t.terratestConfig = config.LoadTFPConfigs(t.cattleConfig)

	adminUser := &management.User{
		Username: "admin",
		Password: t.rancherConfig.AdminPassword,
	}

	t.adminUser = adminUser

	userToken, err := token.GenerateUserToken(adminUser, t.rancherConfig.Host)
	require.NoError(t.T(), err)

	t.rancherConfig.AdminToken = userToken.Token

	client, err := rancher.NewClient(t.rancherConfig.AdminToken, testSession)
	require.NoError(t.T(), err)

	t.client = client
	t.client.RancherConfig.AdminToken = t.rancherConfig.AdminToken

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
		{"RKE1", nodeRolesDedicated, "linode_rke1"},
		{"RKE2", nodeRolesDedicated, "linode_rke2"},
		{"K3S", nodeRolesDedicated, "linode_k3s"},
	}

	for _, tt := range tests {
		cattleConfig := t.TfpSetupSuite()
		configMap := []map[string]any{cattleConfig}

		operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])

		provisioning.GetK8sVersion(t.T(), t.client, t.terratestConfig, t.terraformConfig, configs.DefaultK8sVersion, configMap)

		terratest := new(config.TerratestConfig)
		operations.LoadObjectFromMap(config.TerratestConfigurationFileKey, configMap[0], terratest)

		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion
		testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

		t.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
			defer cleanup.Cleanup(t.T(), t.terraformOptions, keyPath)

			clusterIDs := provisioning.Provision(t.T(), t.client, t.rancherConfig, t.terraformConfig, t.terratestConfig, testUser, testPassword, clusterName, poolName, t.terraformOptions, configMap)
			provisioning.VerifyClustersState(t.T(), t.client, clusterIDs)
			provisioning.VerifyWorkloads(t.T(), t.client, clusterIDs)
		})
	}

	if t.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpSanityTestSuite(t *testing.T) {
	suite.Run(t, new(TfpSanityTestSuite))
}
