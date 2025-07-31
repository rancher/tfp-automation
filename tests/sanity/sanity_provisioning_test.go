package sanity

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

type TfpSanityProvisioningTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	session                    *session.Session
	cattleConfig               map[string]any
	rancherConfig              *rancher.Config
	terraformConfig            *config.TerraformConfig
	terratestConfig            *config.TerratestConfig
	standaloneTerraformOptions *terraform.Options
	terraformOptions           *terraform.Options
}

func (s *TfpSanityProvisioningTestSuite) TearDownSuite() {
	_, keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath, s.terratestConfig.PathToRepo, s.terraformConfig.Provider)
	cleanup.Cleanup(s.T(), s.standaloneTerraformOptions, keyPath)
}

func (s *TfpSanityProvisioningTestSuite) SetupSuite() {
	testSession := session.NewSession()
	s.session = testSession

	s.client, _, s.standaloneTerraformOptions, s.terraformOptions, s.cattleConfig = infrastructure.SetupRancher(s.T(), s.session, keypath.SanityKeyPath)
	s.rancherConfig, s.terraformConfig, s.terratestConfig = config.LoadTFPConfigs(s.cattleConfig)
}

func (s *TfpSanityProvisioningTestSuite) TestTfpProvisioningSanity() {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		module    string
	}{
		{"Sanity RKE2", nodeRolesDedicated, modules.EC2RKE2},
		{"Sanity RKE2 Windows 2019", nil, modules.CustomEC2RKE2Windows2019},
		{"Sanity RKE2 Windows 2022", nil, modules.CustomEC2RKE2Windows2022},
		{"Sanity K3S", nodeRolesDedicated, modules.EC2K3s},
	}

	customClusterNames := []string{}
	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(s.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{s.cattleConfig})
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, s.client.RancherConfig.AdminToken, configMap[0])
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(s.T(), err)

		provisioning.GetK8sVersion(s.T(), s.client, s.terratestConfig, s.terraformConfig, configs.DefaultK8sVersion, configMap)

		rancher, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		currentDate := time.Now().Format("2006-01-02 03:04PM")
		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion + " " + currentDate

		s.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, s.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(s.T(), s.terraformOptions, keyPath)

			clusterIDs, customClusterNames := provisioning.Provision(s.T(), s.client, rancher, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file, false, false, true, customClusterNames)
			provisioning.VerifyClustersState(s.T(), s.client, clusterIDs)

			if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
				clusterIDs, _ := provisioning.Provision(s.T(), s.client, rancher, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file, true, true, true, customClusterNames)
				provisioning.VerifyClustersState(s.T(), s.client, clusterIDs)
			}
		})
	}

	if s.terratestConfig.LocalQaseReporting {
		qase.ReportTest(s.terratestConfig)
	}
}

func TestTfpSanityProvisioningTestSuite(t *testing.T) {
	suite.Run(t, new(TfpSanityProvisioningTestSuite))
}
