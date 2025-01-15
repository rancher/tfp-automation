package airgap

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/token"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/framework"
	airgapCleanup "github.com/rancher/tfp-automation/framework/cleanup/airgap"
	cleanup "github.com/rancher/tfp-automation/framework/cleanup/rancher2"
	resources "github.com/rancher/tfp-automation/framework/set/resources/airgap"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TfpAirgapProvisioningTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	session                    *session.Session
	rancherConfig              *rancher.Config
	terraformConfig            *config.TerraformConfig
	terratestConfig            *config.TerratestConfig
	standaloneTerraformOptions *terraform.Options
	terraformOptions           *terraform.Options
	adminUser                  *management.User
	registry                   string
}

func (a *TfpAirgapProvisioningTestSuite) TearDownSuite() {
	airgapCleanup.ConfigAirgapCleanup(a.T(), a.standaloneTerraformOptions)
}

func (a *TfpAirgapProvisioningTestSuite) SetupSuite() {
	a.terraformConfig = new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, a.terraformConfig)

	a.terratestConfig = new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, a.terratestConfig)

	standaloneTerraformOptions, keyPath := framework.AirgapSetup(a.T(), a.terraformConfig, a.terratestConfig)
	a.standaloneTerraformOptions = standaloneTerraformOptions

	registry, err := resources.CreateMainTF(a.T(), a.standaloneTerraformOptions, keyPath, a.terraformConfig, a.terratestConfig)
	require.NoError(a.T(), err)

	a.registry = registry
}

func (a *TfpAirgapProvisioningTestSuite) TfpSetupSuite(terratestConfig *config.TerratestConfig, terraformConfig *config.TerraformConfig) {
	testSession := session.NewSession()
	a.session = testSession

	rancherConfig := new(rancher.Config)
	ranchFrame.LoadConfig(configs.Rancher, rancherConfig)

	a.rancherConfig = rancherConfig

	adminUser := &management.User{
		Username: "admin",
		Password: rancherConfig.AdminPassword,
	}

	a.adminUser = adminUser

	userToken, err := token.GenerateUserToken(adminUser, a.rancherConfig.Host)
	require.NoError(a.T(), err)

	client, err := rancher.NewClient(userToken.Token, testSession)
	require.NoError(a.T(), err)

	a.client = client

	rancherConfig.AdminToken = userToken.Token

	terraformOptions := framework.Rancher2Setup(a.T(), a.rancherConfig, terraformConfig, terratestConfig)
	a.terraformOptions = terraformOptions
}

func (a *TfpAirgapProvisioningTestSuite) TestTfpAirgapProvisioning() {
	tests := []struct {
		name   string
		module string
	}{
		{"RKE2", "airgap_rke2"},
		{"K3S", "airgap_k3s"},
	}

	for _, tt := range tests {
		terratestConfig := *a.terratestConfig
		terraformConfig := *a.terraformConfig
		terraformConfig.Module = tt.module

		terraformConfig.PrivateRegistries.SystemDefaultRegistry = a.registry
		terraformConfig.PrivateRegistries.URL = a.registry

		a.TfpSetupSuite(&terratestConfig, &terraformConfig)

		provisioning.GetK8sVersion(a.T(), a.client, &terratestConfig, &terraformConfig, configs.DefaultK8sVersion)

		tt.name = tt.name + " Kubernetes version: " + terratestConfig.KubernetesVersion
		testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

		a.Run((tt.name), func() {
			defer cleanup.ConfigCleanup(a.T(), a.terraformOptions)

			provisioning.Provision(a.T(), a.client, a.rancherConfig, &terraformConfig, &terratestConfig, testUser, testPassword, clusterName, poolName, a.terraformOptions, nil)
			provisioning.VerifyCluster(a.T(), a.client, clusterName, &terraformConfig, &terratestConfig)
		})
	}

	if a.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func (a *TfpAirgapProvisioningTestSuite) TestTfpAirgapUpgrading() {
	tests := []struct {
		name   string
		module string
	}{
		{"RKE2", "airgap_rke2"},
		{"K3S", "airgap_k3s"},
	}

	for _, tt := range tests {
		terratestConfig := *a.terratestConfig
		terraformConfig := *a.terraformConfig
		terraformConfig.Module = tt.module

		terraformConfig.PrivateRegistries.SystemDefaultRegistry = a.registry
		terraformConfig.PrivateRegistries.URL = a.registry

		a.TfpSetupSuite(&terratestConfig, &terraformConfig)

		provisioning.GetK8sVersion(a.T(), a.client, &terratestConfig, &terraformConfig, configs.SecondHighestVersion)

		tt.name = tt.name + " Kubernetes version: " + terratestConfig.KubernetesVersion
		testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

		a.Run((tt.name), func() {
			defer cleanup.ConfigCleanup(a.T(), a.terraformOptions)

			provisioning.Provision(a.T(), a.client, a.rancherConfig, &terraformConfig, &terratestConfig, testUser, testPassword, clusterName, poolName, a.terraformOptions, nil)
			provisioning.VerifyCluster(a.T(), a.client, clusterName, &terraformConfig, &terratestConfig)

			provisioning.KubernetesUpgrade(a.T(), a.client, a.rancherConfig, &terraformConfig, &terratestConfig, testUser, testPassword, clusterName, poolName, a.terraformOptions)
			provisioning.VerifyCluster(a.T(), a.client, clusterName, &terraformConfig, &terratestConfig)
		})
	}

	if a.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpAirgapProvisioningTestSuite(t *testing.T) {
	suite.Run(t, new(TfpAirgapProvisioningTestSuite))
}
