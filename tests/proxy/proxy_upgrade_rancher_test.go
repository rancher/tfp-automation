package proxy

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

type TfpProxyUpgradeRancherTestSuite struct {
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
	proxyBastion               string
	proxyPrivateIP             string
}

func (p *TfpProxyUpgradeRancherTestSuite) TearDownSuite() {
	_, keyPath := rancher2.SetKeyPath(keypath.ProxyKeyPath, p.terratestConfig.PathToRepo, p.terraformConfig.Provider)
	cleanup.Cleanup(p.T(), p.standaloneTerraformOptions, keyPath)

	_, keyPath = rancher2.SetKeyPath(keypath.UpgradeKeyPath, p.terratestConfig.PathToRepo, p.terraformConfig.Provider)
	cleanup.Cleanup(p.T(), p.upgradeTerraformOptions, keyPath)
}

func (p *TfpProxyUpgradeRancherTestSuite) SetupSuite() {
	testSession := session.NewSession()
	p.session = testSession

	p.client, p.proxyBastion, p.proxyPrivateIP, p.standaloneTerraformOptions, p.terraformOptions, p.cattleConfig = infrastructure.SetupProxyRancher(p.T(), p.session, keypath.ProxyKeyPath)
	p.rancherConfig, p.terraformConfig, p.terratestConfig, p.standaloneConfig = config.LoadTFPConfigs(p.cattleConfig)

	provisioning.VerifyRancherVersion(p.T(), p.rancherConfig.Host, p.standaloneConfig.RancherTagVersion)
}

func (p *TfpProxyUpgradeRancherTestSuite) TestTfpUpgradeProxyRancher() {
	var clusterIDs []string

	testUser, testPassword := configs.CreateTestCredentials()

	p.provisionAndVerifyCluster("Pre-Upgrade Proxy ", clusterIDs, testUser, testPassword)

	p.client, p.cattleConfig, p.terraformOptions, p.upgradeTerraformOptions = infrastructure.UpgradeProxyRancher(p.T(), p.client, p.proxyPrivateIP, p.proxyBastion, p.session, p.cattleConfig)

	provisioning.VerifyRancherVersion(p.T(), p.rancherConfig.Host, p.standaloneConfig.UpgradedRancherTagVersion)

	p.rancherConfig, p.terraformConfig, p.terratestConfig, _ = config.LoadTFPConfigs(p.cattleConfig)
	p.provisionAndVerifyCluster("Post-Upgrade Proxy ", clusterIDs, testUser, testPassword)

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, p.terratestConfig.PathToRepo, "")
	cleanup.Cleanup(p.T(), p.terraformOptions, keyPath)

	if p.terratestConfig.LocalQaseReporting {
		qase.ReportTest(p.terratestConfig)
	}
}

func (p *TfpProxyUpgradeRancherTestSuite) provisionAndVerifyCluster(name string, clusterIDs []string, testUser, testPassword string) []string {
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

	newFile, rootBody, file := rancher2.InitializeMainTF(p.terratestConfig)
	defer file.Close()

	customClusterNames := []string{}

	for _, tt := range tests {
		configMap, err := provisioning.UniquifyTerraform([]map[string]any{p.cattleConfig})
		require.NoError(p.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, p.client.RancherConfig.AdminToken, configMap[0])
		require.NoError(p.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		require.NoError(p.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(p.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "proxy", "proxyBastion"}, p.proxyBastion, configMap[0])
		require.NoError(p.T(), err)

		provisioning.GetK8sVersion(p.T(), p.client, p.terratestConfig, p.terraformConfig, configs.DefaultK8sVersion, configMap)

		rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

		currentDate := time.Now().Format("2006-01-02 03:04PM")
		tt.name = name + tt.name + " Kubernetes version: " + terratest.KubernetesVersion + " " + currentDate

		p.Run((tt.name), func() {
			clusterIDs, customClusterNames = provisioning.Provision(p.T(), p.client, rancher, terraform, terratest, testUser, testPassword, p.terraformOptions, configMap, newFile, rootBody, file, false, true, true, customClusterNames)
			provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)

			if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
				clusterIDs, customClusterNames = provisioning.Provision(p.T(), p.client, rancher, terraform, terratest, testUser, testPassword, p.terraformOptions, configMap, newFile, rootBody, file, true, true, true, customClusterNames)
				provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)
			}
		})
	}

	return clusterIDs
}

func TestTfpProxyUpgradeRancherTestSuite(t *testing.T) {
	suite.Run(t, new(TfpProxyUpgradeRancherTestSuite))
}
