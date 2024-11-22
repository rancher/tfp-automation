package sanity

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/token"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	standaloneFramework "github.com/rancher/tfp-automation/standalone/framework"
	standaloneCleanup "github.com/rancher/tfp-automation/standalone/framework/cleanup"
	"github.com/rancher/tfp-automation/standalone/framework/resources"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TfpSanityTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	session                    *session.Session
	rancherConfig              *rancher.Config
	terraformConfig            *config.TerraformConfig
	clusterConfig              *config.TerratestConfig
	standaloneTerraformOptions *terraform.Options
	terraformOptions           *terraform.Options
	adminUser                  *management.User
}

func (t *TfpSanityTestSuite) TearDownSuite() {
	standaloneCleanup.StandaloneConfigCleanup(t.T(), t.standaloneTerraformOptions)
}

func (t *TfpSanityTestSuite) SetupSuite() {
	t.terraformConfig = new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, t.terraformConfig)

	t.clusterConfig = new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, t.clusterConfig)

	standaloneTerraformOptions := standaloneFramework.Setup(t.T(), t.terraformConfig)
	t.standaloneTerraformOptions = standaloneTerraformOptions

	resources.CreateMainTF(t.T(), t.standaloneTerraformOptions, t.terraformConfig)
}

func (t *TfpSanityTestSuite) TfpSetupSuite(clusterConfig *config.TerratestConfig, terraformConfig *config.TerraformConfig) {
	testSession := session.NewSession()
	t.session = testSession

	rancherConfig := new(rancher.Config)
	ranchFrame.LoadConfig(configs.Rancher, rancherConfig)

	t.rancherConfig = rancherConfig

	adminUser := &management.User{
		Username: "admin",
		Password: rancherConfig.AdminPassword,
	}

	t.adminUser = adminUser

	userToken, err := token.GenerateUserToken(adminUser, t.rancherConfig.Host)
	require.NoError(t.T(), err)

	client, err := rancher.NewClient(userToken.Token, testSession)
	require.NoError(t.T(), err)

	t.client = client

	rancherConfig.AdminToken = userToken.Token

	terraformOptions := framework.Setup(t.T(), t.rancherConfig, terraformConfig, clusterConfig)
	t.terraformOptions = terraformOptions

	provisioning.GetK8sVersion(t.T(), t.client, clusterConfig, terraformConfig, configs.DefaultK8sVersion)
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
		clusterConfig := *t.clusterConfig
		clusterConfig.Nodepools = tt.nodeRoles
		terraformConfig := *t.terraformConfig
		terraformConfig.Module = tt.module

		t.TfpSetupSuite(&clusterConfig, &terraformConfig)

		tt.name = tt.name + " Kubernetes version: " + clusterConfig.KubernetesVersion
		testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

		t.Run((tt.name), func() {
			defer cleanup.ConfigCleanup(t.T(), t.terraformOptions)

			provisioning.Provision(t.T(), t.client, t.rancherConfig, &terraformConfig, &clusterConfig, testUser, testPassword, clusterName, poolName, t.terraformOptions)
			provisioning.VerifyCluster(t.T(), t.client, clusterName, &terraformConfig, t.terraformOptions, &clusterConfig)
		})
	}

	if t.clusterConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpSanityTestSuite(t *testing.T) {
	suite.Run(t, new(TfpSanityTestSuite))
}
