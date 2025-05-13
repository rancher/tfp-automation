package provisioning

import (
	"os"
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

type ProvisionImportTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	cattleConfig     map[string]any
	rancherConfig    *rancher.Config
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (p *ProvisionImportTestSuite) SetupSuite() map[string]any {
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

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, "")
	terraformOptions := framework.Setup(p.T(), p.terraformConfig, p.terratestConfig, keyPath)
	p.terraformOptions = terraformOptions

	return p.cattleConfig
}

func (p *ProvisionImportTestSuite) TestTfpProvisionImport() {
	tests := []struct {
		name   string
		module string
	}{
		{"Importing TFP RKE1", modules.ImportEC2RKE1},
		{"Importing TFP RKE2", modules.ImportEC2RKE2},
		{"Importing TFP RKE2 Windows", modules.ImportEC2RKE2Windows},
		{"Importing TFP K3S", modules.ImportEC2K3s},
	}

	newFile, rootBody, file := rancher2.InitializeMainTF()
	defer file.Close()

	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		cattleConfig := p.SetupSuite()
		configMap := []map[string]any{cattleConfig}

		_, err := operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(p.T(), err)

		rancher, terraform, _ := config.LoadTFPConfigs(configMap[0])

		p.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, "")
			defer cleanup.Cleanup(p.T(), p.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(p.T(), p.client)
			require.NoError(p.T(), err)

			clusterIDs, _ := provisioning.Provision(p.T(), p.client, rancher, terraform, testUser, testPassword, p.terraformOptions, configMap, newFile, rootBody, file, false, false, true, nil)
			provisioning.VerifyClustersState(p.T(), adminClient, clusterIDs)
		})
	}

	if p.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func (p *ProvisionImportTestSuite) TestTfpProvisionImportDynamicInput() {
	tests := []struct {
		name string
	}{
		{config.StandardClientName.String()},
	}

	newFile, rootBody, file := rancher2.InitializeMainTF()
	defer file.Close()

	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		cattleConfig := p.SetupSuite()
		configMap := []map[string]any{cattleConfig}

		provisioning.GetK8sVersion(p.T(), p.client, p.terratestConfig, p.terraformConfig, configs.DefaultK8sVersion, configMap)

		tt.name = tt.name + " Module: " + p.terraformConfig.Module + " Kubernetes version: " + p.terratestConfig.KubernetesVersion

		p.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, "")
			defer cleanup.Cleanup(p.T(), p.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(p.T(), p.client)
			require.NoError(p.T(), err)

			clusterIDs, _ := provisioning.Provision(p.T(), p.client, p.rancherConfig, p.terraformConfig, testUser, testPassword, p.terraformOptions, configMap, newFile, rootBody, file, false, false, true, nil)
			provisioning.VerifyClustersState(p.T(), adminClient, clusterIDs)
		})
	}

	if p.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpProvisionImportTestSuite(t *testing.T) {
	suite.Run(t, new(ProvisionImportTestSuite))
}
