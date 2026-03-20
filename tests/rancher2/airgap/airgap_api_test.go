//go:build validation || recurring

package airgap

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/extensions/ssh"
	"github.com/rancher/tfp-automation/tests/infrastructure/ranchers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func (a *AirgapAPITestSuite) TearDownSuite() {
	if a.tunnel != nil {
		a.tunnel.StopBastionSSHTunnel()
	}
}

type AirgapAPITestSuite struct {
	suite.Suite
	client             *rancher.Client
	standardUserClient *rancher.Client
	session            *session.Session
	cattleConfig       map[string]any
	rancherConfig      *rancher.Config
	terraformConfig    *config.TerraformConfig
	terratestConfig    *config.TerratestConfig
	standaloneConfig   *config.Standalone
	terraformOptions   *terraform.Options
	tunnel             *ssh.BastionSSHTunnel
}

func (a *AirgapAPITestSuite) SetupSuite() {
	a.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	a.rancherConfig, a.terraformConfig, a.terratestConfig, a.standaloneConfig = config.LoadTFPConfigs(a.cattleConfig)

	testSession := session.NewSession()
	a.session = testSession

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, a.terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(a.T(), a.terraformConfig, a.terratestConfig, keyPath)
	a.terraformOptions = terraformOptions

	sshKey, err := os.ReadFile(a.terraformConfig.PrivateKeyPath)
	require.NoError(a.T(), err)

	a.tunnel, err = ssh.StartBastionSSHTunnel(a.terraformConfig.AirgapBastion, a.standaloneConfig.OSUser, sshKey, "8443", a.standaloneConfig.RancherHostname, "443")
	require.NoError(a.T(), err)

	client, err := ranchers.PostRancherSetup(a.T(), a.terraformOptions, a.rancherConfig, a.session, a.terraformConfig.Standalone.RancherHostname, keyPath, false)
	require.NoError(a.T(), err)

	a.client = client
}

func (a *AirgapAPITestSuite) TestTfpAirgapAPI() {
	tests := []struct {
		name    string
		setting string
	}{
		{"UI_Offline_Preferred_Dynamic", "dynamic"},
		{"UI_Offline_Preferred_Local", "true"},
		{"UI_Offline_Preferred_Remote", "false"},
	}

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(a.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{a.cattleConfig})
		require.NoError(a.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, a.rancherConfig.AdminToken, configMap[0])
		require.NoError(a.T(), err)

		rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

		a.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, a.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(a.T(), a.terraformOptions, keyPath)

			provisioning.AirgapUIOfflinePreferred(a.T(), a.terraformOptions, rancher, terraform, terratest, rootBody, newFile, file, tt.setting, configMap)
		})

		params := tfpQase.GetProvisioningSchemaParams(configMap[0])
		err = qase.UpdateSchemaParameters(tt.name, params)
		if err != nil {
			logrus.Warningf("Failed to upload schema parameters %s", err)
		}
	}

	if a.terratestConfig.LocalQaseReporting {
		results.ReportTest(a.terratestConfig)
	}
}

func TestAirgapAPITestSuite(t *testing.T) {
	suite.Run(t, new(AirgapAPITestSuite))
}
