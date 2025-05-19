package upgrading

import (
	"os"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type KubernetesUpgradeHostedTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	cattleConfig     map[string]any
	rancherConfig    *rancher.Config
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (k *KubernetesUpgradeHostedTestSuite) SetupSuite() {
	testSession := session.NewSession()
	k.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(k.T(), err)

	k.client = client

	k.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	k.rancherConfig, k.terraformConfig, k.terratestConfig = config.LoadTFPConfigs(k.cattleConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, k.terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(k.T(), k.terraformConfig, k.terratestConfig, keyPath)
	k.terraformOptions = terraformOptions
}

func (k *KubernetesUpgradeHostedTestSuite) TestTfpKubernetesUpgradeHosted() {
	tests := []struct {
		name string
	}{
		{config.StandardClientName.String()},
	}

	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(k.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{k.cattleConfig})
		require.NoError(k.T(), err)

		tt.name = tt.name + " Module: " + k.terraformConfig.Module + " Kubernetes version: " + k.terratestConfig.KubernetesVersion

		k.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, k.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(k.T(), k.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(k.T(), k.client)
			require.NoError(k.T(), err)

			clusterIDs, _ := provisioning.Provision(k.T(), k.client, k.rancherConfig, k.terraformConfig, k.terratestConfig, testUser, testPassword, k.terraformOptions, configMap, newFile, rootBody, file, false, false, false, nil)
			provisioning.VerifyClustersState(k.T(), adminClient, clusterIDs)

			provisioning.KubernetesUpgrade(k.T(), k.client, k.rancherConfig, k.terraformConfig, k.terratestConfig, testUser, testPassword, k.terraformOptions, configMap, newFile, rootBody, file, false)

			time.Sleep(4 * time.Minute)

			provisioning.VerifyClustersState(k.T(), adminClient, clusterIDs)
			provisioning.VerifyKubernetesVersion(k.T(), k.client, clusterIDs[0], k.terratestConfig.KubernetesVersion, k.terraformConfig.Module)
		})
	}

	if k.terratestConfig.LocalQaseReporting {
		qase.ReportTest(k.terratestConfig)
	}
}

func TestTfpKubernetesUpgradeHostedTestSuite(t *testing.T) {
	suite.Run(t, new(KubernetesUpgradeHostedTestSuite))
}
