package psact

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/framework"
	cleanup "github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PSACTTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	terraformConfig  *config.TerraformConfig
	clusterConfig    *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (p *PSACTTestSuite) SetupSuite() {
	testSession := session.NewSession()
	p.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(p.T(), err)

	p.client = client

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, terraformConfig)

	p.terraformConfig = terraformConfig

	clusterConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, clusterConfig)

	p.clusterConfig = clusterConfig

	terraformOptions := framework.Setup(p.T())
	p.terraformOptions = terraformOptions

	provisioning.GetK8sVersion(p.T(), p.client, p.clusterConfig, p.terraformConfig, configs.DefaultK8sVersion)
}

func (p *PSACTTestSuite) TestTfpPSACT() {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		psact     config.PSACT
	}{
		{
			name:      "Rancher Privileged " + config.StandardClientName.String(),
			nodeRoles: nodeRolesDedicated,
			psact:     "rancher-privileged",
		},
		{
			name:      "Rancher Restricted " + config.StandardClientName.String(),
			nodeRoles: nodeRolesDedicated,
			psact:     "rancher-restricted",
		},
		{
			name:      "Rancher Baseline " + config.StandardClientName.String(),
			nodeRoles: nodeRolesDedicated,
			psact:     "rancher-baseline",
		},
	}

	for _, tt := range tests {
		clusterConfig := *p.clusterConfig
		clusterConfig.Nodepools = tt.nodeRoles
		clusterConfig.PSACT = string(tt.psact)

		tt.name = tt.name + " Module: " + p.terraformConfig.Module + " Kubernetes version: " + p.clusterConfig.KubernetesVersion

		clusterName := namegen.AppendRandomString(configs.TFP)
		poolName := namegen.AppendRandomString(configs.TFP)

		p.Run((tt.name), func() {
			defer cleanup.Cleanup(p.T(), p.terraformOptions)

			provisioning.Provision(p.T(), clusterName, poolName, &clusterConfig, p.terraformOptions)
			provisioning.VerifyCluster(p.T(), p.client, clusterName, p.terraformConfig, p.terraformOptions, &clusterConfig)
		})
	}
}

func TestTfpPSACTTestSuite(t *testing.T) {
	suite.Run(t, new(PSACTTestSuite))
}
