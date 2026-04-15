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
	clusterActions "github.com/rancher/tests/actions/clusters"
	provisioningActions "github.com/rancher/tests/actions/provisioning"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	nested "github.com/rancher/tfp-automation/tests/extensions/nestedModules"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/extensions/ssh"
	"github.com/rancher/tfp-automation/tests/infrastructure/ranchers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AirgapACETestSuite struct {
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

func (a *AirgapACETestSuite) TearDownSuite() {
	if a.tunnel != nil {
		a.tunnel.StopBastionSSHTunnel()
	}
}

func (a *AirgapACETestSuite) SetupSuite() {
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

func (a *AirgapACETestSuite) TestTfpAirgapACE() {
	var err error
	var testUser, testPassword string
	var clusterIDs []string

	customClusterNames := []string{}

	a.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(a.client)
	require.NoError(a.T(), err)

	standardUserToken, err := ranchers.CreateStandardUserToken(a.T(), a.terraformOptions, a.rancherConfig, testUser, testPassword)
	require.NoError(a.T(), err)

	standardToken := standardUserToken.Token

	localAuthEndpoint := config.TerraformConfig{
		LocalAuthEndpoint: true,
	}

	tests := []struct {
		name         string
		module       string
		authEndpoint config.TerraformConfig
	}{
		{"RKE2_Airgap_ACE", modules.AirgapAWSRKE2, localAuthEndpoint},
		{"K3S_Airgap_ACE", modules.AirgapAWSK3S, localAuthEndpoint},
	}

	for _, tt := range tests {
		a.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(a.terraformConfig, a.terratestConfig, a.terraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()

			cattleConfig, err := provisioning.UniquifyTerraform(a.cattleConfig)
			require.NoError(t, err)

			_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, cattleConfig)
			require.NoError(a.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "localAuthEndpoint"}, tt.authEndpoint.LocalAuthEndpoint, cattleConfig)
			require.NoError(a.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, cattleConfig)
			require.NoError(a.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, a.terraformConfig.PrivateRegistries.SystemDefaultRegistry, cattleConfig)
			require.NoError(a.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, a.terraformConfig.PrivateRegistries.URL, cattleConfig)
			require.NoError(a.T(), err)

			provisioning.GetK8sVersion(a.client, cattleConfig)

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(cattleConfig)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, a.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(a.T(), perTestTerraformOptions, keyPath)

			clusters, _ := provisioning.Provision(a.T(), a.client, a.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, newFile, rootBody, file, false, false, true, clusterIDs, customClusterNames, nestedRancherModuleDir)
			err = provisioningActions.VerifyClusterReady(a.client, clusters[0])
			require.NoError(a.T(), err)

			err = clusterActions.VerifyServiceAccountTokenSecret(a.client, clusters[0].Name)
			require.NoError(a.T(), err)

			err = pods.VerifyClusterPods(a.client, clusters[0])
			require.NoError(a.T(), err)

			provisioning.VerifyACEAirgap(a.T(), a.client, clusters[0])

			params := tfpQase.GetProvisioningSchemaParams(cattleConfig)
			err = qase.UpdateSchemaParameters(tt.name, params)
			if err != nil {
				logrus.Warningf("Failed to upload schema parameters %s", err)
			}
		})
	}

	if a.terratestConfig.LocalQaseReporting {
		results.ReportTest(a.terratestConfig)
	}
}

func TestAirgapACETestSuite(t *testing.T) {
	suite.Run(t, new(AirgapACETestSuite))
}
