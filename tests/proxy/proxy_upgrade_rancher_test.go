package proxy

import (
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
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
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	nested "github.com/rancher/tfp-automation/tests/extensions/nestedModules"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/infrastructure/ranchers"
	"github.com/sirupsen/logrus"
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

	p.client, p.proxyBastion, p.proxyPrivateIP, p.standaloneTerraformOptions, p.terraformOptions, p.cattleConfig = ranchers.SetupProxyRancher(p.T(), p.session, keypath.ProxyKeyPath)
	p.rancherConfig, p.terraformConfig, p.terratestConfig, p.standaloneConfig = config.LoadTFPConfigs(p.cattleConfig)
}

func (p *TfpProxyUpgradeRancherTestSuite) TestTfpUpgradeProxyRancher() {
	standardUserClient, standardToken, testUser, testPassword := ranchers.SetupResources(p.T(), p.client, p.rancherConfig, p.terratestConfig, p.terraformOptions)

	p.rancherConfig, p.terraformConfig, p.terratestConfig, _ = config.LoadTFPConfigs(p.cattleConfig)
	nestedRancherModuleDir := p.provisionAndVerifyCluster("Proxy_Pre_Rancher_Upgrade", standardUserClient, standardToken, testUser, testPassword)

	p.client, p.cattleConfig, p.terraformOptions, p.upgradeTerraformOptions = ranchers.UpgradeProxyRancher(p.T(), p.client, p.proxyPrivateIP, p.proxyBastion, p.session, p.cattleConfig)

	ranchers.CleanupDownstreamClusters(p.T(), p.client, p.terraformConfig)
	os.RemoveAll(nestedRancherModuleDir)

	standardUserClient, standardToken, testUser, testPassword = ranchers.SetupResources(p.T(), p.client, p.rancherConfig, p.terratestConfig, p.terraformOptions)

	p.rancherConfig, p.terraformConfig, p.terratestConfig, _ = config.LoadTFPConfigs(p.cattleConfig)
	nestedRancherModuleDir = p.provisionAndVerifyCluster("Proxy_Post_Rancher_Upgrade", standardUserClient, standardToken, testUser, testPassword)

	ranchers.CleanupDownstreamClusters(p.T(), p.client, p.terraformConfig)
	os.RemoveAll(nestedRancherModuleDir)

	if p.terratestConfig.LocalQaseReporting {
		results.ReportTest(p.terratestConfig)
	}
}

func (p *TfpProxyUpgradeRancherTestSuite) provisionAndVerifyCluster(name string, standardUserClient *rancher.Client, standardToken,
	testUser, testPassword string) string {
	var clusterIDs []string
	var nestedRancherModuleDir string

	customClusterNames := []string{}
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}
	nodeRolesWindows := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool, config.WindowsNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		module    string
	}{
		{name + "_RKE2", nodeRolesDedicated, modules.NodeDriverAWSRKE2},
		{name + "_RKE2_Windows_2019", nodeRolesWindows, modules.CustomAWSRKE2Windows2019},
		{name + "_RKE2_Windows_2022", nodeRolesWindows, modules.CustomAWSRKE2Windows2022},
		{name + "_K3S", nodeRolesDedicated, modules.NodeDriverAWSK3S},
	}

	p.T().Run(name, func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(p.terraformConfig, p.terratestConfig, p.terraformOptions, tt.name, configs.NestedRancherModuleDir)
				require.NoError(t, err)

				newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
				defer file.Close()

				cattleConfig, err := provisioning.UniquifyTerraform(p.cattleConfig)
				require.NoError(t, err)

				_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, cattleConfig)
				require.NoError(p.T(), err)

				_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, cattleConfig)
				require.NoError(p.T(), err)

				_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, cattleConfig)
				require.NoError(p.T(), err)

				_, err = operations.ReplaceValue([]string{"terraform", "proxy", "proxyBastion"}, p.proxyBastion, cattleConfig)
				require.NoError(p.T(), err)

				provisioning.GetK8sVersion(standardUserClient, cattleConfig)

				rancher, terraform, terratest, _ := config.LoadTFPConfigs(cattleConfig)

				tt.name = name + tt.name

				clusters, customClusterNames := provisioning.Provision(p.T(), p.client, standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, newFile, rootBody, file, false, true, true, clusterIDs, customClusterNames, nestedRancherModuleDir)
				err = provisioningActions.VerifyClusterReady(p.client, clusters[0])
				require.NoError(p.T(), err)

				err = clusterActions.VerifyServiceAccountTokenSecret(p.client, clusters[0].Name)
				require.NoError(p.T(), err)

				err = pods.VerifyClusterPods(p.client, clusters[0])
				require.NoError(p.T(), err)

				if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
					clusters, customClusterNames = provisioning.Provision(p.T(), p.client, standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, newFile, rootBody, file, true, true, true, clusterIDs, customClusterNames, nestedRancherModuleDir)
					err = provisioningActions.VerifyClusterReady(p.client, clusters[0])
					require.NoError(p.T(), err)

					err = clusterActions.VerifyServiceAccountTokenSecret(p.client, clusters[0].Name)
					require.NoError(p.T(), err)

					err = pods.VerifyClusterPods(p.client, clusters[0])
					require.NoError(p.T(), err)
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

func TestTfpProxyUpgradeRancherTestSuite(t *testing.T) {
	suite.Run(t, new(TfpProxyUpgradeRancherTestSuite))
}
