package tests

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
	"github.com/josh-diamond/tfp-automation/config"
	"github.com/josh-diamond/tfp-automation/framework"
	cleanup "github.com/josh-diamond/tfp-automation/framework/cleanup"
	"github.com/josh-diamond/tfp-automation/tests/extensions/provisioning"
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

func (r *ScaleTestSuite) SetupSuite() {
	testSession := session.NewSession()
	r.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(r.T(), err)

	r.client = client

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
	require.NoError(r.T(), err)

	newUser.Password = user.Password

	standardUserClient, err := client.AsUser(newUser)
	require.NoError(r.T(), err)

	r.standardUserClient = standardUserClient

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig("terraform", terraformConfig)

	r.terraformConfig = terraformConfig

	clusterConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig("terratest", clusterConfig)

	r.clusterConfig = clusterConfig

	terraformOptions := framework.Setup(r.T())
	r.terraformOptions = terraformOptions
}

func (r *ScaleTestSuite) TestScale() {
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
		{"Scaling 1 node all roles -> 4 nodes -> 3 nodes " + config.StandardClientName.String(), nodeRolesAll, scaleUpRolesAll, scaleDownRolesAll, r.standardUserClient},
		{"Scaling 2 nodes shared roles -> 6 nodes -> 4 nodes  " + config.StandardClientName.String(), nodeRolesShared, scaleUpRolesShared, scaleDownRolesShared, r.standardUserClient},
		{"Scaling 3 nodes dedicated roles -> 8 nodes -> 6 nodes " + config.StandardClientName.String(), nodeRolesDedicated, scaleUpRolesDedicated, scaleDownRolesDedicated, r.standardUserClient},
	}

	for _, tt := range tests {
		clusterConfig := *r.clusterConfig
		clusterConfig.Nodepools = tt.nodeRoles
		clusterConfig.ScaledUpNodepools = tt.scaleUpNodeRoles
		clusterConfig.ScaledDownNodepools = tt.scaleDownNodeRoles

		var scaledUpCount, scaledDownCount int64

		for _, scaleUpNodepool := range tt.scaleUpNodeRoles {
			scaledUpCount += scaleUpNodepool.Quantity
		}

		for _, scaleDownNodepool := range tt.scaleDownNodeRoles {
			scaledDownCount += scaleDownNodepool.Quantity
		}

		clusterName := namegen.AppendRandomString(provisioning.TFP)

		r.Run((tt.name), func() {
			provisioning.Provision(r.T(), clusterName, r.terraformConfig, &clusterConfig, r.terraformOptions)
			provisioning.VerifyCluster(r.T(), tt.client, clusterName, r.terraformConfig, r.terraformOptions, &clusterConfig)

			provisioning.ScaleUp(r.T(), clusterName, r.terraformOptions, &clusterConfig)
			provisioning.VerifyCluster(r.T(), tt.client, clusterName, r.terraformConfig, r.terraformOptions, &clusterConfig)
			provisioning.VerifyNodeCount(r.T(), tt.client, clusterName, scaledUpCount)

			provisioning.ScaleDown(r.T(), clusterName, r.terraformOptions, &clusterConfig)
			provisioning.VerifyCluster(r.T(), tt.client, clusterName, r.terraformConfig, r.terraformOptions, &clusterConfig)
			provisioning.VerifyNodeCount(r.T(), tt.client, clusterName, scaledDownCount)
			cleanup.Cleanup(r.T(), r.terraformOptions)
		})
	}
}

func (r *ScaleTestSuite) TestScaleDynamicInput() {
	tests := []struct {
		name   string
		client *rancher.Client
	}{
		{config.StandardClientName.String(), r.standardUserClient},
	}

	for _, tt := range tests {
		clusterName := namegen.AppendRandomString(provisioning.TFP)

		r.Run((tt.name), func() {
			provisioning.Provision(r.T(), clusterName, r.terraformConfig, r.clusterConfig, r.terraformOptions)
			provisioning.VerifyCluster(r.T(), r.client, clusterName, r.terraformConfig, r.terraformOptions, r.clusterConfig)

			provisioning.ScaleUp(r.T(), clusterName, r.terraformOptions, r.clusterConfig)
			provisioning.VerifyCluster(r.T(), r.client, clusterName, r.terraformConfig, r.terraformOptions, r.clusterConfig)
			provisioning.VerifyNodeCount(r.T(), r.client, clusterName, r.clusterConfig.ScaledUpNodeCount)

			provisioning.ScaleDown(r.T(), clusterName, r.terraformOptions, r.clusterConfig)
			provisioning.VerifyCluster(r.T(), r.client, clusterName, r.terraformConfig, r.terraformOptions, r.clusterConfig)
			provisioning.VerifyNodeCount(r.T(), r.client, clusterName, r.clusterConfig.ScaledDownNodeCount)
			cleanup.Cleanup(r.T(), r.terraformOptions)
		})
	}
}

func TestScaleTestSuite(t *testing.T) {
	suite.Run(t, new(ScaleTestSuite))
}
