package proxy

import (
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/token"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/pipeline"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
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
	proxyPrivateIP             string
}

func (p *TfpProxyUpgradeRancherTestSuite) TearDownSuite() {
	keyPath := rancher2.SetKeyPath(keypath.ProxyKeyPath, p.terraformConfig.Provider)
	cleanup.Cleanup(p.T(), p.standaloneTerraformOptions, keyPath)

	keyPath = rancher2.SetKeyPath(keypath.UpgradeKeyPath, p.terraformConfig.Provider)
	cleanup.Cleanup(p.T(), p.upgradeTerraformOptions, keyPath)
}

func (p *TfpProxyUpgradeRancherTestSuite) SetupSuite() {
	p.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	p.rancherConfig, p.terraformConfig, p.terratestConfig = config.LoadTFPConfigs(p.cattleConfig)

	keyPath := rancher2.SetKeyPath(keypath.ProxyKeyPath, p.terraformConfig.Provider)
	standaloneTerraformOptions := framework.Setup(p.T(), p.terraformConfig, p.terratestConfig, keyPath)
	p.standaloneTerraformOptions = standaloneTerraformOptions

	proxyNode, proxyPrivateIP, err := resources.CreateMainTF(p.T(), p.standaloneTerraformOptions, keyPath, p.terraformConfig, p.terratestConfig)
	require.NoError(p.T(), err)

	p.proxyNode = proxyNode
	p.proxyPrivateIP = proxyPrivateIP

	keyPath = rancher2.SetKeyPath(keypath.UpgradeKeyPath, p.terraformConfig.Provider)
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

	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, "")
	terraformOptions := framework.Setup(p.T(), p.terraformConfig, p.terratestConfig, keyPath)
	p.terraformOptions = terraformOptions

	return p.cattleConfig
}

func (p *TfpProxyUpgradeRancherTestSuite) TestTfpUpgradeProxyRancher() {
	var clusterIDs []string

	p.provisionAndVerifyCluster("Pre-Upgrade Proxy ", clusterIDs, false)

	p.terraformConfig.Standalone.UpgradeProxyRancher = true

	keyPath := rancher2.SetKeyPath(keypath.UpgradeKeyPath, p.terraformConfig.Provider)
	err := upgrade.CreateMainTF(p.T(), p.upgradeTerraformOptions, keyPath, p.terraformConfig, p.terratestConfig, p.proxyPrivateIP, p.proxyNode, "", "")
	require.NoError(p.T(), err)

	provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)

	p.provisionAndVerifyCluster("Post-Upgrade Proxy ", clusterIDs, true)

	if p.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func (p *TfpProxyUpgradeRancherTestSuite) provisionAndVerifyCluster(name string, clusterIDs []string, deleteClusters bool) []string {
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
		cattleConfig := p.TfpSetupSuite()
		configMap := []map[string]any{cattleConfig}

		_, err := operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		require.NoError(p.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(p.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "proxy", "proxyBastion"}, p.proxyNode, configMap[0])
		require.NoError(p.T(), err)

		provisioning.GetK8sVersion(p.T(), p.client, p.terratestConfig, p.terraformConfig, configs.DefaultK8sVersion, configMap)

		rancher, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		tt.name = name + tt.name + " Kubernetes version: " + terratest.KubernetesVersion

		p.Run((tt.name), func() {
			clusterIDs, customClusterNames = provisioning.Provision(p.T(), p.client, rancher, terraform, testUser, testPassword, p.terraformOptions, configMap, newFile, rootBody, file, false, true, true, customClusterNames)
			provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)

			if strings.Contains(terraform.Module, modules.CustomEC2RKE2Windows) {
				clusterIDs, _ = provisioning.Provision(p.T(), p.client, rancher, terraform, testUser, testPassword, p.terraformOptions, configMap, newFile, rootBody, file, true, true, true, customClusterNames)
				provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)
			}
		})
	}

	if deleteClusters {
		keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, "")
		cleanup.Cleanup(p.T(), p.terraformOptions, keyPath)
	}

	return clusterIDs
}

func TestTfpProxyUpgradeRancherTestSuite(t *testing.T) {
	suite.Run(t, new(TfpProxyUpgradeRancherTestSuite))
}
