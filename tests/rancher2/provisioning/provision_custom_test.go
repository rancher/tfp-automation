package provisioning

import (
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
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

type ProvisionCustomTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	cattleConfig     map[string]any
	rancherConfig    *rancher.Config
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (p *ProvisionCustomTestSuite) SetupSuite() map[string]any {
	testSession := session.NewSession()
	p.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(p.T(), err)

	p.client = client

	p.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	configMap, err := provisioning.UniquifyTerraform([]map[string]any{p.cattleConfig})
	require.NoError(p.T(), err)

	p.cattleConfig = configMap[0]
	p.rancherConfig, p.terraformConfig, p.terratestConfig = config.LoadTFPConfigs(p.cattleConfig)

	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, nil)
	terraformOptions := framework.Setup(p.T(), p.terraformConfig, p.terratestConfig, keyPath)
	p.terraformOptions = terraformOptions

	return p.cattleConfig
}

func (p *ProvisionCustomTestSuite) TestTfpProvisionCustom() {
	tests := []struct {
		name   string
		module string
	}{
		{"Custom TFP RKE1", "ec2_rke1_custom"},
		{"Custom TFP RKE2", "ec2_rke2_custom"},
		{"Custom TFP RKE2 Windows", "ec2_rke2_windows_custom"},
		{"Custom TFP K3S", "ec2_k3s_custom"},
	}

	for _, tt := range tests {
		cattleConfig := p.SetupSuite()
		configMap := []map[string]any{cattleConfig}

		operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])

		provisioning.GetK8sVersion(p.T(), p.client, p.terratestConfig, p.terraformConfig, configs.DefaultK8sVersion, configMap)

		_, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion
		testUser, testPassword := configs.CreateTestCredentials()

		p.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, nil)
			defer cleanup.Cleanup(p.T(), p.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(p.T(), p.client)
			require.NoError(p.T(), err)

			clusterIDs := provisioning.Provision(p.T(), p.client, p.rancherConfig, terraform, terratest, testUser, testPassword, p.terraformOptions, configMap, false)
			provisioning.VerifyClustersState(p.T(), adminClient, clusterIDs)

			if strings.Contains(terraform.Module, modules.CustomEC2RKE2Windows) {
				clusterIDs = provisioning.Provision(p.T(), p.client, p.rancherConfig, terraform, terratest, testUser, testPassword, p.terraformOptions, configMap, true)
				provisioning.VerifyClustersState(p.T(), adminClient, clusterIDs)
			}
		})
	}

	if p.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpProvisionCustomTestSuite(t *testing.T) {
	suite.Run(t, new(ProvisionCustomTestSuite))
}
