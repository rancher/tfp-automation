package psact

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/framework"
	cleanup "github.com/rancher/tfp-automation/framework/cleanup/rancher2"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PSACTTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	rancherConfig    *rancher.Config
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (p *PSACTTestSuite) SetupSuite() {
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

	terraformOptions := framework.Rancher2Setup(p.T(), p.rancherConfig, p.terraformConfig, p.terratestConfig)
	p.terraformOptions = terraformOptions

	provisioning.GetK8sVersion(p.T(), p.client, p.terratestConfig, p.terraformConfig, configs.DefaultK8sVersion)
}

func (p *PSACTTestSuite) TestTfpPSACT() {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		psact     config.PSACT
	}{
		{"Rancher Privileged " + config.StandardClientName.String(), nodeRolesDedicated, "rancher-privileged"},
		{"Rancher Restricted " + config.StandardClientName.String(), nodeRolesDedicated, "rancher-restricted"},
		{"Rancher Baseline " + config.StandardClientName.String(), nodeRolesDedicated, "rancher-baseline"},
	}

	for _, tt := range tests {
		terratestConfig := *p.terratestConfig
		terratestConfig.Nodepools = tt.nodeRoles
		terratestConfig.PSACT = string(tt.psact)

		tt.name = tt.name + " Module: " + p.terraformConfig.Module + " Kubernetes version: " + p.terratestConfig.KubernetesVersion

		testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

		p.Run((tt.name), func() {
			defer cleanup.ConfigCleanup(p.T(), p.terraformOptions)

			adminClient, err := provisioning.FetchAdminClient(p.T(), p.client)
			require.NoError(p.T(), err)

			provisioning.Provision(p.T(), p.client, p.rancherConfig, p.terraformConfig, &terratestConfig, testUser, testPassword, clusterName, poolName, p.terraformOptions, nil)
			provisioning.VerifyCluster(p.T(), adminClient, clusterName, p.terraformConfig, p.terraformOptions, &terratestConfig)
		})
	}

	if p.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpPSACTTestSuite(t *testing.T) {
	suite.Run(t, new(PSACTTestSuite))
}
