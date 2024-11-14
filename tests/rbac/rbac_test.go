package rbac

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	rb "github.com/rancher/tfp-automation/tests/extensions/rbac"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RBACTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	rancherConfig    *rancher.Config
	terraformConfig  *config.TerraformConfig
	clusterConfig    *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (r *RBACTestSuite) SetupSuite() {
	testSession := session.NewSession()
	r.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(r.T(), err)

	r.client = client

	rancherConfig := new(rancher.Config)
	ranchFrame.LoadConfig(configs.Rancher, rancherConfig)

	r.rancherConfig = rancherConfig

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, terraformConfig)

	r.terraformConfig = terraformConfig

	clusterConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, clusterConfig)

	r.clusterConfig = clusterConfig

	terraformOptions := framework.Setup(r.T(), r.rancherConfig, r.terraformConfig, r.clusterConfig)
	r.terraformOptions = terraformOptions

	provisioning.GetK8sVersion(r.T(), r.client, r.clusterConfig, r.terraformConfig, configs.DefaultK8sVersion)
}

func (r *RBACTestSuite) TestTfpRBAC() {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name     string
		rbacRole config.Role
	}{
		{"Cluster Owner", config.ClusterOwner},
		{"Project Owner", config.ProjectOwner},
	}

	for _, tt := range tests {
		tt.name = tt.name + " Module: " + r.terraformConfig.Module

		r.Run((tt.name), func() {
			defer cleanup.ConfigCleanup(r.T(), r.terraformOptions)

			clusterConfig := *r.clusterConfig
			clusterConfig.Nodepools = nodeRolesDedicated

			testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

			provisioning.Provision(r.T(), r.client, r.rancherConfig, r.terraformConfig, &clusterConfig, testUser, testPassword, clusterName, poolName, r.terraformOptions)
			provisioning.VerifyCluster(r.T(), r.client, clusterName, r.terraformConfig, r.terraformOptions, &clusterConfig)

			rb.RBAC(r.T(), r.client, r.rancherConfig, r.terraformConfig, &clusterConfig, testUser, testPassword, clusterName, poolName, r.terraformOptions, tt.rbacRole)
		})
	}

	if r.clusterConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpRBACTestSuite(t *testing.T) {
	suite.Run(t, new(RBACTestSuite))
}
