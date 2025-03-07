package proxy

import (
	"os"
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
	cattleConfig               map[string]any
	rancherConfig              *rancher.Config
	terraformConfig            *config.TerraformConfig
	terratestConfig            *config.TerratestConfig
	standaloneTerraformOptions *terraform.Options
	upgradeTerraformOptions    *terraform.Options
	terraformOptions           *terraform.Options
	proxyNode                  string
	proxyServerNodeOne         string
}

func (p *TfpProxyUpgradeRancherTestSuite) TearDownSuite() {
	keyPath := rancher2.SetKeyPath(keypath.ProxyKeyPath)
	cleanup.Cleanup(p.T(), p.standaloneTerraformOptions, keyPath)

	keyPath = rancher2.SetKeyPath(keypath.UpgradeKeyPath)
	cleanup.Cleanup(p.T(), p.upgradeTerraformOptions, keyPath)
}

func (p *TfpProxyUpgradeRancherTestSuite) SetupSuite() {
	p.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	p.rancherConfig, p.terraformConfig, p.terratestConfig = config.LoadTFPConfigs(p.cattleConfig)

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

func (p *TfpProxyUpgradeRancherTestSuite) TfpSetupSuite() map[string]any {
	testSession := session.NewSession()
	p.session = testSession

	p.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	configMap, err := provisioning.UniquifyTerraform([]map[string]any{p.cattleConfig})
	require.NoError(p.T(), err)

	p.cattleConfig = configMap[0]
	p.rancherConfig, p.terraformConfig, p.terratestConfig = config.LoadTFPConfigs(p.cattleConfig)

	adminUser := &management.User{
		Username: "admin",
		Password: p.rancherConfig.AdminPassword,
	}

	userToken, err := token.GenerateUserToken(adminUser, p.rancherConfig.Host)
	require.NoError(p.T(), err)

	p.rancherConfig.AdminToken = userToken.Token

	client, err := rancher.NewClient(p.rancherConfig.AdminToken, testSession)
	require.NoError(p.T(), err)

	p.client = client
	p.client.RancherConfig.AdminToken = p.rancherConfig.AdminToken
	p.client.RancherConfig.AdminPassword = p.rancherConfig.AdminPassword
	p.client.RancherConfig.Host = p.rancherConfig.Host

	operations.ReplaceValue([]string{"rancher", "adminToken"}, p.rancherConfig.AdminToken, configMap[0])
	operations.ReplaceValue([]string{"rancher", "adminPassword"}, p.rancherConfig.AdminPassword, configMap[0])
	operations.ReplaceValue([]string{"rancher", "host"}, p.rancherConfig.Host, configMap[0])

	err = pipeline.PostRancherInstall(p.client, p.client.RancherConfig.AdminPassword)
	require.NoError(p.T(), err)

	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
	terraformOptions := framework.Setup(p.T(), p.terraformConfig, p.terratestConfig, keyPath)
	p.terraformOptions = terraformOptions

	return p.cattleConfig
}

func (p *TfpProxyUpgradeRancherTestSuite) TestTfpUpgradeProxyRancher() {
	p.provisionAndVerifyCluster("Pre-Upgrade Proxy ")

	p.terraformConfig.Standalone.UpgradeProxyRancher = true

	keyPath := rancher2.SetKeyPath(keypath.UpgradeKeyPath)
	err := upgrade.CreateMainTF(p.T(), p.upgradeTerraformOptions, keyPath, p.terraformConfig, p.terratestConfig, p.proxyServerNodeOne, p.proxyNode, "", "")
	require.NoError(p.T(), err)

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
		cattleConfig := p.TfpSetupSuite()
		configMap := []map[string]any{cattleConfig}

		operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		operations.ReplaceValue([]string{"terraform", "proxy", "proxyBastion"}, p.proxyNode, configMap[0])

		provisioning.GetK8sVersion(p.T(), p.client, p.terratestConfig, p.terraformConfig, configs.DefaultK8sVersion, configMap)

		terraform := new(config.TerraformConfig)
		operations.LoadObjectFromMap(config.TerraformConfigurationFileKey, configMap[0], terraform)

		terratest := new(config.TerratestConfig)
		operations.LoadObjectFromMap(config.TerratestConfigurationFileKey, configMap[0], terratest)

		tt.name = name + tt.name + " Kubernetes version: " + terratest.KubernetesVersion
		testUser, testPassword := configs.CreateTestCredentials()

		p.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
			defer cleanup.Cleanup(p.T(), p.terraformOptions, keyPath)

			clusterIDs := provisioning.Provision(p.T(), p.client, p.rancherConfig, terraform, terratest, testUser, testPassword, p.terraformOptions, configMap, false)
			provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)
		})
	}
}

func TestTfpProxyUpgradeRancherTestSuite(t *testing.T) {
	suite.Run(t, new(TfpProxyUpgradeRancherTestSuite))
}
