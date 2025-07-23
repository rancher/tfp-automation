package proxy

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	resources "github.com/rancher/tfp-automation/framework/set/resources/proxy"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/infrastructure"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TfpProxyProvisioningTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	session                    *session.Session
	cattleConfig               map[string]any
	rancherConfig              *rancher.Config
	terraformConfig            *config.TerraformConfig
	terratestConfig            *config.TerratestConfig
	standaloneTerraformOptions *terraform.Options
	terraformOptions           *terraform.Options
	proxyBastion               string
}

func (p *TfpProxyProvisioningTestSuite) TearDownSuite() {
	_, keyPath := rancher2.SetKeyPath(keypath.ProxyKeyPath, p.terratestConfig.PathToRepo, p.terraformConfig.Provider)
	cleanup.Cleanup(p.T(), p.standaloneTerraformOptions, keyPath)
}

func (p *TfpProxyProvisioningTestSuite) SetupSuite() {
	p.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	p.rancherConfig, p.terraformConfig, p.terratestConfig = config.LoadTFPConfigs(p.cattleConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.ProxyKeyPath, p.terratestConfig.PathToRepo, p.terraformConfig.Provider)
	standaloneTerraformOptions := framework.Setup(p.T(), p.terraformConfig, p.terratestConfig, keyPath)
	p.standaloneTerraformOptions = standaloneTerraformOptions

	proxyBastion, _, err := resources.CreateMainTF(p.T(), p.standaloneTerraformOptions, keyPath, p.rancherConfig, p.terraformConfig, p.terratestConfig)
	require.NoError(p.T(), err)

	p.proxyBastion = proxyBastion

	testSession := session.NewSession()
	p.session = testSession

	client, err := infrastructure.PostRancherSetup(p.T(), p.rancherConfig, testSession, p.terraformConfig.Standalone.RancherHostname, false, false)
	if err != nil && *p.rancherConfig.Cleanup {
		cleanup.Cleanup(p.T(), p.standaloneTerraformOptions, keyPath)
	}

	p.client = client
	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, p.terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(p.T(), p.terraformConfig, p.terratestConfig, keyPath)
	p.terraformOptions = terraformOptions
}

func (p *TfpProxyProvisioningTestSuite) TestTfpNoProxyProvisioning() {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		module    string
	}{
		{"No Proxy RKE2", nodeRolesDedicated, modules.EC2RKE2},
		{"No Proxy RKE2 Windows 2019", nil, modules.CustomEC2RKE2Windows2019},
		{"No Proxy RKE2 Windows 2022", nil, modules.CustomEC2RKE2Windows2022},
		{"No Proxy K3S", nodeRolesDedicated, modules.EC2K3s},
	}

	customClusterNames := []string{}
	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(p.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{p.cattleConfig})
		require.NoError(p.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, p.client.RancherConfig.AdminToken, configMap[0])
		require.NoError(p.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		require.NoError(p.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(p.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "proxy", "proxyBastion"}, "", configMap[0])
		require.NoError(p.T(), err)

		provisioning.GetK8sVersion(p.T(), p.client, p.terratestConfig, p.terraformConfig, configs.DefaultK8sVersion, configMap)

		rancher, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		currentDate := time.Now().Format("2006-01-02 03:04PM")
		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion + " " + currentDate

		p.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, p.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(p.T(), p.terraformOptions, keyPath)

			clusterIDs, customClusterNames := provisioning.Provision(p.T(), p.client, rancher, terraform, terratest, testUser, testPassword, p.terraformOptions, configMap, newFile, rootBody, file, false, false, true, customClusterNames)
			provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)

			if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
				clusterIDs, _ := provisioning.Provision(p.T(), p.client, rancher, terraform, terratest, testUser, testPassword, p.terraformOptions, configMap, newFile, rootBody, file, true, true, true, customClusterNames)
				provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)
			}
		})
	}

	if p.terratestConfig.LocalQaseReporting {
		qase.ReportTest(p.terratestConfig)
	}
}

func (p *TfpProxyProvisioningTestSuite) TestTfpProxyProvisioning() {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		module    string
	}{
		{"Proxy RKE2", nodeRolesDedicated, modules.EC2RKE2},
		{"Proxy RKE2 Windows 2019", nil, modules.CustomEC2RKE2Windows2019},
		{"Proxy RKE2 Windows 2022", nil, modules.CustomEC2RKE2Windows2022},
		{"Proxy K3S", nodeRolesDedicated, modules.EC2K3s},
	}

	customClusterNames := []string{}
	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(p.terratestConfig)
		defer file.Close()

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

		rancher, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		currentDate := time.Now().Format("2006-01-02 03:04PM")
		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion + " " + currentDate

		p.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, p.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(p.T(), p.terraformOptions, keyPath)

			clusterIDs, _ := provisioning.Provision(p.T(), p.client, rancher, terraform, terratest, testUser, testPassword, p.terraformOptions, configMap, newFile, rootBody, file, false, false, true, customClusterNames)
			provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)

			if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
				clusterIDs, _ := provisioning.Provision(p.T(), p.client, rancher, terraform, terratest, testUser, testPassword, p.terraformOptions, configMap, newFile, rootBody, file, true, true, true, customClusterNames)
				provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)
			}
		})
	}

	if p.terratestConfig.LocalQaseReporting {
		qase.ReportTest(p.terratestConfig)
	}
}

func TestTfpProxyProvisioningTestSuite(t *testing.T) {
	suite.Run(t, new(TfpProxyProvisioningTestSuite))
}
