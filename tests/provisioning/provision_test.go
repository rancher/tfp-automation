package tests

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/framework"
	cleanup "github.com/rancher/tfp-automation/framework/cleanup"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ProvisionTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	rancherConfig    *rancher.Config
	terraformConfig  *config.TerraformConfig
	clusterConfig    *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (p *ProvisionTestSuite) SetupSuite() {
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

	clusterConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, clusterConfig)

	p.clusterConfig = clusterConfig

	terraformOptions := framework.Setup(p.T(), p.rancherConfig, p.terraformConfig, p.clusterConfig)
	p.terraformOptions = terraformOptions

	provisioning.GetK8sVersion(p.T(), p.client, p.clusterConfig, p.terraformConfig, configs.DefaultK8sVersion)
}

func (p *ProvisionTestSuite) TestTfpProvision() {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
	}{
		{"3 nodes - 1 role per node " + config.StandardClientName.String(), nodeRolesDedicated},
	}

	for _, tt := range tests {
		clusterConfig := *p.clusterConfig
		clusterConfig.Nodepools = tt.nodeRoles

		tt.name = tt.name + " Module: " + p.terraformConfig.Module + " Kubernetes version: " +
			p.clusterConfig.KubernetesVersion

		testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

		p.Run((tt.name), func() {
			defer cleanup.ConfigCleanup(p.T(), p.terraformOptions)

			adminClient, err := provisioning.FetchAdminClient(p.T(), p.client)
			require.NoError(p.T(), err)

			provisioning.Provision(p.T(), p.client, p.rancherConfig, p.terraformConfig, &clusterConfig, testUser, testPassword, clusterName, poolName, p.terraformOptions, nil)
			provisioning.VerifyCluster(p.T(), adminClient, clusterName, p.terraformConfig, p.terraformOptions, &clusterConfig)
		})
	}

	if p.clusterConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func (p *ProvisionTestSuite) TestTfpProvisionDynamicInput() {
	tests := []struct {
		name string
	}{
		{config.StandardClientName.String()},
	}

	for _, tt := range tests {
		tt.name = tt.name + " Module: " + p.terraformConfig.Module + " Kubernetes version: " + p.clusterConfig.KubernetesVersion

		testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

		p.Run((tt.name), func() {
			defer cleanup.ConfigCleanup(p.T(), p.terraformOptions)

			adminClient, err := provisioning.FetchAdminClient(p.T(), p.client)
			require.NoError(p.T(), err)

			provisioning.Provision(p.T(), p.client, p.rancherConfig, p.terraformConfig, p.clusterConfig, testUser, testPassword, clusterName, poolName, p.terraformOptions, nil)
			provisioning.VerifyCluster(p.T(), adminClient, clusterName, p.terraformConfig, p.terraformOptions, p.clusterConfig)
		})
	}

	if p.clusterConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpProvisionTestSuite(t *testing.T) {
	suite.Run(t, new(ProvisionTestSuite))
}
