package tests

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

type ProvisionTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
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

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, terraformConfig)

	p.terraformConfig = terraformConfig

	clusterConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, clusterConfig)

	p.clusterConfig = clusterConfig

	terraformOptions := framework.Setup(p.T())
	p.terraformOptions = terraformOptions

	provisioning.DefaultK8sVersion(p.T(), p.client, p.clusterConfig, p.terraformConfig)
}

func (p *ProvisionTestSuite) TestTfpProvision() {
	nodeRolesAll := []config.Nodepool{config.AllRolesNodePool}
	nodeRolesShared := []config.Nodepool{config.EtcdControlPlaneNodePool, config.WorkerNodePool}
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
	}{
		{"1 Node all roles " + config.StandardClientName.String(), nodeRolesAll},
		{"2 nodes - etcd|cp roles per 1 node " + config.StandardClientName.String(), nodeRolesShared},
		{"3 nodes - 1 role per node " + config.StandardClientName.String(), nodeRolesDedicated},
	}

	for _, tt := range tests {
		clusterConfig := *p.clusterConfig
		clusterConfig.Nodepools = tt.nodeRoles

		tt.name = tt.name + " Module: " + p.terraformConfig.Module + " Kubernetes version: " +
			p.clusterConfig.KubernetesVersion

		clusterName := namegen.AppendRandomString(configs.TFP)
		poolName := namegen.AppendRandomString(configs.TFP)

		p.Run((tt.name), func() {
			defer cleanup.Cleanup(p.T(), p.terraformOptions)

			provisioning.Provision(p.T(), clusterName, poolName, &clusterConfig, p.terraformOptions)
			provisioning.VerifyCluster(p.T(), p.client, clusterName, p.terraformConfig, p.terraformOptions, &clusterConfig)
		})
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

		clusterName := namegen.AppendRandomString(configs.TFP)
		poolName := namegen.AppendRandomString(configs.TFP)

		p.Run((tt.name), func() {
			defer cleanup.Cleanup(p.T(), p.terraformOptions)

			provisioning.Provision(p.T(), clusterName, poolName, p.clusterConfig, p.terraformOptions)
			provisioning.VerifyCluster(p.T(), p.client, clusterName, p.terraformConfig, p.terraformOptions, p.clusterConfig)
		})
	}
}

func TestTfpProvisionTestSuite(t *testing.T) {
	suite.Run(t, new(ProvisionTestSuite))
}
