package provisioning

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
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

type ProvisionHostedTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	rancherConfig    *rancher.Config
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (p *ProvisionHostedTestSuite) SetupSuite() {
	testSession := session.NewSession()
	p.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(p.T(), err)

	p.client = client

	rancherConfig := new(rancher.Config)
	ranchFrame.LoadConfig(configs.Rancher, rancherConfig)

	p.rancherConfig = rancherConfig

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, terraformConfig)

	p.terraformConfig = terraformConfig

	terratestConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, terratestConfig)

	p.terratestConfig = terratestConfig

	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
	terraformOptions := framework.Setup(p.T(), p.terraformConfig, p.terratestConfig, keyPath)
	p.terraformOptions = terraformOptions
}

func (p *ProvisionHostedTestSuite) TestTfpProvisionHosted() {
	tests := []struct {
		name string
	}{
		{config.StandardClientName.String()},
	}

	for _, tt := range tests {
		tt.name = tt.name + " Module: " + p.terraformConfig.Module + " Kubernetes version: " + p.terratestConfig.KubernetesVersion

		testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

		p.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
			defer cleanup.Cleanup(p.T(), p.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(p.T(), p.client)
			require.NoError(p.T(), err)

			clusterIDs := provisioning.Provision(p.T(), p.client, p.rancherConfig, p.terraformConfig, p.terratestConfig, testUser, testPassword, clusterName, poolName, p.terraformOptions, nil)
			provisioning.VerifyClustersState(p.T(), adminClient, clusterIDs)
			provisioning.VerifyWorkloads(p.T(), adminClient, clusterIDs)
			provisioning.VerifyKubernetesVersion(p.T(), adminClient, clusterIDs[0], p.terratestConfig.KubernetesVersion, p.terraformConfig.Module)
		})
	}

	if p.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpProvisionHostedTestSuite(t *testing.T) {
	suite.Run(t, new(ProvisionHostedTestSuite))
}
