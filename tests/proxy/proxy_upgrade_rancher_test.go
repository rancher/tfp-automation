package proxy

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/infrastructure/ranchers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TfpProxyUpgradeRancherTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	standardUserClient         *rancher.Client
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
	var clusterIDs []string

	p.rancherConfig, p.terraformConfig, p.terratestConfig, _ = config.LoadTFPConfigs(p.cattleConfig)
	p.provisionAndVerifyCluster("Proxy_Pre_Rancher_Upgrade_", clusterIDs)

	p.client, p.cattleConfig, p.terraformOptions, p.upgradeTerraformOptions = ranchers.UpgradeProxyRancher(p.T(), p.client, p.proxyPrivateIP, p.proxyBastion, p.session, p.cattleConfig)

	p.rancherConfig, p.terraformConfig, p.terratestConfig, _ = config.LoadTFPConfigs(p.cattleConfig)
	p.provisionAndVerifyCluster("Proxy_Post_Rancher_Upgrade_", clusterIDs)

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, p.terratestConfig.PathToRepo, "")
	cleanup.Cleanup(p.T(), p.terraformOptions, keyPath)

	if p.terratestConfig.LocalQaseReporting {
		results.ReportTest(p.terratestConfig)
	}
}

func (p *TfpProxyUpgradeRancherTestSuite) provisionAndVerifyCluster(name string, clusterIDs []string) []string {
	var err error
	var testUser, testPassword string

	newFile, rootBody, file := rancher2.InitializeMainTF(p.terratestConfig)
	defer file.Close()

	customClusterNames := []string{}

	p.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(p.client)
	require.NoError(p.T(), err)

	standardUserToken, err := ranchers.CreateStandardUserToken(p.T(), p.terraformOptions, p.rancherConfig, testUser, testPassword)
	require.NoError(p.T(), err)

	standardToken := standardUserToken.Token

	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		module    string
	}{
		{"RKE2", nodeRolesDedicated, modules.EC2RKE2},
		{"RKE2_Windows_2019", nil, modules.CustomEC2RKE2Windows2019},
		{"RKE2_Windows_2022", nil, modules.CustomEC2RKE2Windows2022},
		{"K3S", nodeRolesDedicated, modules.EC2K3s},
	}

	for _, tt := range tests {
		configMap, err := provisioning.UniquifyTerraform([]map[string]any{p.cattleConfig})
		require.NoError(p.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, configMap[0])
		require.NoError(p.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		require.NoError(p.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(p.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "proxy", "proxyBastion"}, p.proxyBastion, configMap[0])
		require.NoError(p.T(), err)

		provisioning.GetK8sVersion(p.T(), p.standardUserClient, p.terratestConfig, p.terraformConfig, configMap)

		rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

		tt.name = name + tt.name

		p.Run((tt.name), func() {
			clusterIDs, customClusterNames = provisioning.Provision(p.T(), p.client, p.standardUserClient, rancher, terraform, terratest, testUser, testPassword, p.terraformOptions, configMap, newFile, rootBody, file, false, true, true, customClusterNames)
			provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)
			provisioning.VerifyServiceAccountTokenSecret(p.T(), p.client, clusterIDs)

			cluster, err := p.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraform.ResourcePrefix)
			require.NoError(p.T(), err)

			err = pods.VerifyClusterPods(p.client, cluster)
			require.NoError(p.T(), err)

			if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
				clusterIDs, customClusterNames = provisioning.Provision(p.T(), p.client, p.standardUserClient, rancher, terraform, terratest, testUser, testPassword, p.terraformOptions, configMap, newFile, rootBody, file, true, true, true, customClusterNames)
				provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)
				provisioning.VerifyServiceAccountTokenSecret(p.T(), p.client, clusterIDs)
				err = pods.VerifyClusterPods(p.client, cluster)
				require.NoError(p.T(), err)
			}
		})

		params := tfpQase.GetProvisioningSchemaParams(configMap[0])
		err = qase.UpdateSchemaParameters(tt.name, params)
		if err != nil {
			logrus.Warningf("Failed to upload schema parameters %s", err)
		}
	}

	return clusterIDs
}

func TestTfpProxyUpgradeRancherTestSuite(t *testing.T) {
	suite.Run(t, new(TfpProxyUpgradeRancherTestSuite))
}
