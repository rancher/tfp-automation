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
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
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

type TfpProxyProvisioningTestSuite struct {
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
	terraformOptions           *terraform.Options
	proxyBastion               string
}

func (p *TfpProxyProvisioningTestSuite) TearDownSuite() {
	_, keyPath := rancher2.SetKeyPath(keypath.ProxyKeyPath, p.terratestConfig.PathToRepo, p.terraformConfig.Provider)
	cleanup.Cleanup(p.T(), p.standaloneTerraformOptions, keyPath)
}

func (p *TfpProxyProvisioningTestSuite) SetupSuite() {
	testSession := session.NewSession()
	p.session = testSession

	p.client, p.proxyBastion, _, p.standaloneTerraformOptions, p.terraformOptions, p.cattleConfig = ranchers.SetupProxyRancher(p.T(), p.session, keypath.ProxyKeyPath)
	p.rancherConfig, p.terraformConfig, p.terratestConfig, p.standaloneConfig = config.LoadTFPConfigs(p.cattleConfig)
}

func (p *TfpProxyProvisioningTestSuite) TestTfpNoProxyProvisioning() {
	var err error
	var testUser, testPassword string
	var clusterIDs []string

	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		module    string
	}{
		{"No_Proxy_RKE2", nodeRolesDedicated, modules.EC2RKE2},
		{"No_Proxy_RKE2_Windows_2019", nil, modules.CustomEC2RKE2Windows2019},
		{"No_Proxy_RKE2_Windows_2022", nil, modules.CustomEC2RKE2Windows2022},
		{"No_Proxy_K3S", nodeRolesDedicated, modules.EC2K3s},
	}

	customClusterNames := []string{}

	p.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(p.client)
	require.NoError(p.T(), err)

	standardUserToken, err := ranchers.CreateStandardUserToken(p.T(), p.terraformOptions, p.rancherConfig, testUser, testPassword)
	require.NoError(p.T(), err)

	standardToken := standardUserToken.Token

	for _, tt := range tests {
		p.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(p.terraformConfig, p.terratestConfig, p.terraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

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

			_, err = operations.ReplaceValue([]string{"terraform", "proxy", "proxyBastion"}, "", cattleConfig)
			require.NoError(p.T(), err)

			provisioning.GetK8sVersion(p.standardUserClient, cattleConfig)

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(cattleConfig)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, p.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(p.T(), perTestTerraformOptions, keyPath)

			clusters, customClusterNames := provisioning.Provision(p.T(), p.client, p.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, newFile, rootBody, file, false, false, true, clusterIDs, customClusterNames, nestedRancherModuleDir)
			err = provisioningActions.VerifyClusterReady(p.client, clusters[0])
			require.NoError(p.T(), err)

			err = clusterActions.VerifyServiceAccountTokenSecret(p.client, clusters[0].Name)
			require.NoError(p.T(), err)

			err = pods.VerifyClusterPods(p.client, clusters[0])
			require.NoError(p.T(), err)

			if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
				clusters, _ = provisioning.Provision(p.T(), p.client, p.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, newFile, rootBody, file, true, true, true, clusterIDs, customClusterNames, nestedRancherModuleDir)
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

	if p.terratestConfig.LocalQaseReporting {
		results.ReportTest(p.terratestConfig)
	}
}

func (p *TfpProxyProvisioningTestSuite) TestTfpProxyProvisioning() {
	var err error
	var testUser, testPassword string
	var clusterIDs []string

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
		{"Proxy_Provisioning_RKE2", nodeRolesDedicated, modules.EC2RKE2},
		{"Proxy_Provisioning_RKE2_Windows_2019", nil, modules.CustomEC2RKE2Windows2019},
		{"Proxy_Provisioning_RKE2_Windows_2022", nil, modules.CustomEC2RKE2Windows2022},
		{"Proxy_Provisioning_K3S", nodeRolesDedicated, modules.EC2K3s},
	}

	for _, tt := range tests {
		p.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(p.terraformConfig, p.terratestConfig, p.terraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

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

			provisioning.GetK8sVersion(p.standardUserClient, cattleConfig)

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(cattleConfig)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, p.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(p.T(), perTestTerraformOptions, keyPath)

			clusters, _ := provisioning.Provision(p.T(), p.client, p.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, newFile, rootBody, file, false, false, true, clusterIDs, customClusterNames, nestedRancherModuleDir)
			err = provisioningActions.VerifyClusterReady(p.client, clusters[0])
			require.NoError(p.T(), err)

			err = clusterActions.VerifyServiceAccountTokenSecret(p.client, clusters[0].Name)
			require.NoError(p.T(), err)

			err = pods.VerifyClusterPods(p.client, clusters[0])
			require.NoError(p.T(), err)

			if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
				clusters, _ = provisioning.Provision(p.T(), p.client, p.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, newFile, rootBody, file, true, true, true, clusterIDs, customClusterNames, nestedRancherModuleDir)
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

	if p.terratestConfig.LocalQaseReporting {
		results.ReportTest(p.terratestConfig)
	}
}

func TestTfpProxyProvisioningTestSuite(t *testing.T) {
	suite.Run(t, new(TfpProxyProvisioningTestSuite))
}
