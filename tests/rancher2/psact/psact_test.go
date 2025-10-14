//go:build validation || recurring

package psact

import (
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PSACTTestSuite struct {
	suite.Suite
	client             *rancher.Client
	standardUserClient *rancher.Client
	session            *session.Session
	cattleConfig       map[string]any
	rancherConfig      *rancher.Config
	terraformConfig    *config.TerraformConfig
	terratestConfig    *config.TerratestConfig
	terraformOptions   *terraform.Options
}

func (p *PSACTTestSuite) SetupSuite() {
	testSession := session.NewSession()
	p.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(p.T(), err)

	p.client = client

	p.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	p.rancherConfig, p.terraformConfig, p.terratestConfig, _ = config.LoadTFPConfigs(p.cattleConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, p.terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(p.T(), p.terraformConfig, p.terratestConfig, keyPath)
	p.terraformOptions = terraformOptions
}

func (p *PSACTTestSuite) TestTfpPSACT() {
	if p.terraformConfig.Standalone != nil && strings.Contains(p.terraformConfig.Standalone.RancherTagVersion, "2.11") {
		p.T().Skip("Skipping PSACT tests on Rancher v2.11.x due to known issues.")
	}

	var err error
	var testUser, testPassword string

	p.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(p.client)
	require.NoError(p.T(), err)

	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		module    string
		nodeRoles []config.Nodepool
		psact     config.PSACT
	}{
		{"RKE2_Rancher_Privileged", modules.EC2RKE2, nodeRolesDedicated, "rancher-privileged"},
		{"RKE2_Rancher_Restricted", modules.EC2RKE2, nodeRolesDedicated, "rancher-restricted"},
		{"RKE2_Rancher_Baseline", modules.EC2RKE2, nodeRolesDedicated, "rancher-baseline"},
		{"K3S_Rancher_Privileged", modules.EC2K3s, nodeRolesDedicated, "rancher-privileged"},
		{"K3S_Rancher_Restricted", modules.EC2K3s, nodeRolesDedicated, "rancher-restricted"},
		{"K3S_Rancher_Baseline", modules.EC2K3s, nodeRolesDedicated, "rancher-baseline"},
	}

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(p.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{p.cattleConfig})
		require.NoError(p.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(p.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		require.NoError(p.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "psact"}, tt.psact, configMap[0])
		require.NoError(p.T(), err)

		provisioning.GetK8sVersion(p.T(), p.client, p.terratestConfig, p.terraformConfig, configs.DefaultK8sVersion, configMap)

		rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

		p.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, p.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(p.T(), p.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(p.T(), p.client)
			require.NoError(p.T(), err)

			clusterIDs, _ := provisioning.Provision(p.T(), p.client, p.standardUserClient, rancher, terraform, terratest, testUser, testPassword, p.terraformOptions, configMap, newFile, rootBody, file, false, false, false, nil)
			provisioning.VerifyClustersState(p.T(), adminClient, clusterIDs)
			provisioning.VerifyClusterPSACT(p.T(), adminClient, clusterIDs)
		})
	}

	if p.terratestConfig.LocalQaseReporting {
		qase.ReportTest(p.terratestConfig)
	}
}

func TestTfpPSACTTestSuite(t *testing.T) {
	suite.Run(t, new(PSACTTestSuite))
}
