package proxy

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
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	resources "github.com/rancher/tfp-automation/framework/set/resources/proxy"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	upgrade "github.com/rancher/tfp-automation/framework/set/resources/upgrade"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TfpProxyUpgradeRancherTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	session                    *session.Session
	rancherConfig              *rancher.Config
	terraformConfig            *config.TerraformConfig
	terratestConfig            *config.TerratestConfig
	standaloneTerraformOptions *terraform.Options
	upgradeTerraformOptions    *terraform.Options
	terraformOptions           *terraform.Options
	adminUser                  *management.User
	proxyNode                  string
	proxyServerNodeOne         string
}

func (p *TfpProxyUpgradeRancherTestSuite) TearDownSuite() {
	keyPath := rancher2.SetKeyPath(keypath.ProxyKeyPath)
	cleanup.Cleanup(p.T(), p.standaloneTerraformOptions, keyPath)
}

func (p *TfpProxyUpgradeRancherTestSuite) SetupSuite() {
	p.terraformConfig = new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, p.terraformConfig)

	p.terratestConfig = new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, p.terratestConfig)

	keyPath := rancher2.SetKeyPath(keypath.ProxyKeyPath)
	standaloneTerraformOptions := framework.Setup(p.T(), p.terraformConfig, p.terratestConfig, keyPath)
	p.standaloneTerraformOptions = standaloneTerraformOptions

	proxyNode, proxyServerNodeOne, err := resources.CreateMainTF(p.T(), p.standaloneTerraformOptions, keyPath, p.terraformConfig, p.terratestConfig)
	require.NoError(p.T(), err)

	p.proxyNode = proxyNode
	p.proxyServerNodeOne = proxyServerNodeOne

	keyPath = rancher2.SetKeyPath(keypath.UpgradeKeyPath)
	upgradeTerraformOptions := framework.Setup(p.T(), p.terraformConfig, p.terratestConfig, keyPath)

	p.upgradeTerraformOptions = upgradeTerraformOptions
}

func (p *TfpProxyUpgradeRancherTestSuite) TfpSetupSuite(terratestConfig *config.TerratestConfig, terraformConfig *config.TerraformConfig) {
	testSession := session.NewSession()
	p.session = testSession

	rancherConfig := new(rancher.Config)
	ranchFrame.LoadConfig(configs.Rancher, rancherConfig)

	p.rancherConfig = rancherConfig

	adminUser := &management.User{
		Username: "admin",
		Password: rancherConfig.AdminPassword,
	}

	p.adminUser = adminUser

	userToken, err := token.GenerateUserToken(adminUser, p.rancherConfig.Host)
	require.NoError(p.T(), err)

	rancherConfig.AdminToken = userToken.Token

	client, err := rancher.NewClient(rancherConfig.AdminToken, testSession)
	require.NoError(p.T(), err)

	p.client = client
	p.client.RancherConfig.AdminToken = rancherConfig.AdminToken

	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
	terraformOptions := framework.Setup(p.T(), terraformConfig, terratestConfig, keyPath)
	p.terraformOptions = terraformOptions
}

func (p *TfpProxyUpgradeRancherTestSuite) TestTfpUpgradeProxyProvisioning() {
	p.provisionAndVerifyCluster("Pre-Upgrade Proxy ")

	keyPath := rancher2.SetKeyPath(keypath.UpgradeKeyPath)
	upgrade.CreateMainTF(p.T(), p.upgradeTerraformOptions, keyPath, p.terraformConfig, p.terratestConfig, p.proxyNode, p.proxyServerNodeOne)

	p.provisionAndVerifyCluster("Post-Upgrade Proxy ")

	if p.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func (p *TfpProxyUpgradeRancherTestSuite) provisionAndVerifyCluster(name string) {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		module    string
	}{
		{"RKE1", nodeRolesDedicated, "ec2_rke1"},
		{"RKE2", nodeRolesDedicated, "ec2_rke2"},
		{"K3S", nodeRolesDedicated, "ec2_k3s"},
	}

	for _, tt := range tests {
		terratestConfig := *p.terratestConfig
		terratestConfig.Nodepools = tt.nodeRoles
		terraformConfig := *p.terraformConfig
		terraformConfig.Module = tt.module
		terraformConfig.Proxy.ProxyBastion = p.proxyNode

		p.TfpSetupSuite(&terratestConfig, &terraformConfig)

		provisioning.GetK8sVersion(p.T(), p.client, &terratestConfig, &terraformConfig, configs.DefaultK8sVersion)

		tt.name = name + tt.name + " Kubernetes version: " + terratestConfig.KubernetesVersion
		testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

		p.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
			defer cleanup.Cleanup(p.T(), p.terraformOptions, keyPath)

			clusterIDs := provisioning.Provision(p.T(), p.client, p.rancherConfig, &terraformConfig, &terratestConfig, testUser, testPassword, clusterName, poolName, p.terraformOptions, nil)
			provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)
		})
	}
}

func TestTfpProxyUpgradeRancherTestSuite(t *testing.T) {
	suite.Run(t, new(TfpProxyUpgradeRancherTestSuite))
}
