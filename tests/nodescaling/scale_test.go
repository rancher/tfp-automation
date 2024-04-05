package nodescaling

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/users"
	password "github.com/rancher/shepherd/extensions/users/passwordgenerator"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework"
	cleanup "github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ScaleTestSuite struct {
	suite.Suite
	client             *rancher.Client
	standardUserClient *rancher.Client
	session            *session.Session
	terraformConfig    *config.TerraformConfig
	clusterConfig      *config.TerratestConfig
	terraformOptions   *terraform.Options
}

func (s *ScaleTestSuite) SetupSuite() {
	testSession := session.NewSession()
	s.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(s.T(), err)

	s.client = client

	enabled := true
	var testuser = namegen.AppendRandomString("testuser-")
	var testpassword = password.GenerateUserPassword("testpass-")
	user := &management.User{
		Username: testuser,
		Password: testpassword,
		Name:     testuser,
		Enabled:  &enabled,
	}

	newUser, err := users.CreateUserWithRole(client, user, "user")
	require.NoError(s.T(), err)

	newUser.Password = user.Password

	standardUserClient, err := client.AsUser(newUser)
	require.NoError(s.T(), err)

	s.standardUserClient = standardUserClient

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, terraformConfig)

	s.terraformConfig = terraformConfig

	clusterConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, clusterConfig)

	s.clusterConfig = clusterConfig

	terraformOptions := framework.Setup(s.T())
	s.terraformOptions = terraformOptions

	provisioning.DefaultK8sVersion(s.T(), s.client, s.clusterConfig, s.terraformConfig)
}

func (s *ScaleTestSuite) TestTfpScale() {
	nodeRolesAll := []config.Nodepool{config.AllRolesNodePool}
	scaleUpRolesAll := []config.Nodepool{config.ScaleUpAllRolesNodePool}
	scaleDownRolesAll := []config.Nodepool{config.ScaleDownAllRolesNodePool}

	nodeRolesShared := []config.Nodepool{config.EtcdControlPlaneNodePool, config.WorkerNodePool}
	scaleUpRolesShared := []config.Nodepool{config.ScaleUpEtcdControlPlaneNodePool, config.ScaleUpWorkerNodePool}
	scaleDownRolesShared := []config.Nodepool{config.ScaleUpEtcdControlPlaneNodePool, config.WorkerNodePool}

	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}
	scaleUpRolesDedicated := []config.Nodepool{config.ScaleUpEtcdNodePool, config.ScaleUpControlPlaneNodePool, config.ScaleUpWorkerNodePool}
	scaleDownRolesDedicated := []config.Nodepool{config.ScaleUpEtcdNodePool, config.ScaleUpControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name               string
		nodeRoles          []config.Nodepool
		scaleUpNodeRoles   []config.Nodepool
		scaleDownNodeRoles []config.Nodepool
		client             *rancher.Client
	}{
		{"Scaling 1 node all roles -> 4 nodes -> 3 nodes " + config.StandardClientName.String(), nodeRolesAll, scaleUpRolesAll, scaleDownRolesAll, s.standardUserClient},
		{"Scaling 2 nodes shared roles -> 6 nodes -> 4 nodes  " + config.StandardClientName.String(), nodeRolesShared, scaleUpRolesShared, scaleDownRolesShared, s.standardUserClient},
		{"Scaling 3 nodes dedicated roles -> 8 nodes -> 6 nodes " + config.StandardClientName.String(), nodeRolesDedicated, scaleUpRolesDedicated, scaleDownRolesDedicated, s.standardUserClient},
	}

	for _, tt := range tests {
		clusterConfig := *s.clusterConfig
		clusterConfig.Nodepools = tt.nodeRoles
		clusterConfig.ScalingInput.ScaledUpNodepools = tt.scaleUpNodeRoles
		clusterConfig.ScalingInput.ScaledDownNodepools = tt.scaleDownNodeRoles

		var scaledUpCount, scaledDownCount int64

		for _, scaleUpNodepool := range tt.scaleUpNodeRoles {
			scaledUpCount += scaleUpNodepool.Quantity
		}

		for _, scaleDownNodepool := range tt.scaleDownNodeRoles {
			scaledDownCount += scaleDownNodepool.Quantity
		}

		tt.name = tt.name + " Module: " + s.terraformConfig.Module + " Kubernetes version: " +
			s.clusterConfig.KubernetesVersion

		clusterName := namegen.AppendRandomString(provisioning.TFP)
		poolName := namegen.AppendRandomString(provisioning.TFP)

		s.Run((tt.name), func() {
			provisioning.Provision(s.T(), tt.client, clusterName, poolName, &clusterConfig, s.terraformOptions)
			provisioning.VerifyCluster(s.T(), tt.client, clusterName, s.terraformConfig, s.terraformOptions, &clusterConfig)

			clusterConfig.Nodepools = clusterConfig.ScalingInput.ScaledUpNodepools

			provisioning.Scale(s.T(), clusterName, poolName, s.terraformOptions, &clusterConfig)
			provisioning.VerifyCluster(s.T(), tt.client, clusterName, s.terraformConfig, s.terraformOptions, &clusterConfig)
			provisioning.VerifyNodeCount(s.T(), tt.client, clusterName, s.terraformConfig, scaledUpCount)

			clusterConfig.Nodepools = clusterConfig.ScalingInput.ScaledDownNodepools

			provisioning.Scale(s.T(), clusterName, poolName, s.terraformOptions, &clusterConfig)
			provisioning.VerifyCluster(s.T(), tt.client, clusterName, s.terraformConfig, s.terraformOptions, &clusterConfig)
			provisioning.VerifyNodeCount(s.T(), tt.client, clusterName, s.terraformConfig, scaledDownCount)

			cleanup.Cleanup(s.T(), s.terraformOptions)
		})
	}
}

func (s *ScaleTestSuite) TestTfpScaleDynamicInput() {
	tests := []struct {
		name   string
		client *rancher.Client
	}{
		{config.StandardClientName.String(), s.standardUserClient},
	}

	for _, tt := range tests {
		tt.name = tt.name + " Module: " + s.terraformConfig.Module + " Kubernetes version: " + s.clusterConfig.KubernetesVersion

		clusterName := namegen.AppendRandomString(provisioning.TFP)
		poolName := namegen.AppendRandomString(provisioning.TFP)

		s.Run((tt.name), func() {
			defer cleanup.Cleanup(s.T(), s.terraformOptions)

			provisioning.Provision(s.T(), tt.client, clusterName, poolName, s.clusterConfig, s.terraformOptions)
			provisioning.VerifyCluster(s.T(), s.client, clusterName, s.terraformConfig, s.terraformOptions, s.clusterConfig)

			s.clusterConfig.Nodepools = s.clusterConfig.ScalingInput.ScaledUpNodepools

			provisioning.Scale(s.T(), clusterName, poolName, s.terraformOptions, s.clusterConfig)
			provisioning.VerifyCluster(s.T(), s.client, clusterName, s.terraformConfig, s.terraformOptions, s.clusterConfig)
			provisioning.VerifyNodeCount(s.T(), s.client, clusterName, s.terraformConfig, s.clusterConfig.ScalingInput.ScaledUpNodeCount)

			s.clusterConfig.Nodepools = s.clusterConfig.ScalingInput.ScaledDownNodepools

			provisioning.Scale(s.T(), clusterName, poolName, s.terraformOptions, s.clusterConfig)
			provisioning.VerifyCluster(s.T(), s.client, clusterName, s.terraformConfig, s.terraformOptions, s.clusterConfig)
			provisioning.VerifyNodeCount(s.T(), s.client, clusterName, s.terraformConfig, s.clusterConfig.ScalingInput.ScaledDownNodeCount)
		})
	}
}

func TestTfpScaleTestSuite(t *testing.T) {
	suite.Run(t, new(ScaleTestSuite))
}
