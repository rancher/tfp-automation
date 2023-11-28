package tests

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/rancher/tests/framework/clients/rancher"
	management "github.com/rancher/rancher/tests/framework/clients/rancher/generated/management/v3"
	"github.com/rancher/rancher/tests/framework/extensions/users"
	password "github.com/rancher/rancher/tests/framework/extensions/users/passwordgenerator"
	ranchFrame "github.com/rancher/rancher/tests/framework/pkg/config"
	namegen "github.com/rancher/rancher/tests/framework/pkg/namegenerator"
	"github.com/rancher/rancher/tests/framework/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework"
	cleanup "github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ProvisionTestSuite struct {
	suite.Suite
	client             *rancher.Client
	standardUserClient *rancher.Client
	session            *session.Session
	terraformConfig    *config.TerraformConfig
	clusterConfig      *config.TerratestConfig
	terraformOptions   *terraform.Options
}

func (r *ProvisionTestSuite) SetupSuite() {
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

func (r *ProvisionTestSuite) TestProvision() {
	nodeRolesAll := []config.Nodepool{config.AllRolesNodePool}
	nodeRolesShared := []config.Nodepool{config.EtcdControlPlaneNodePool, config.WorkerNodePool}
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		client    *rancher.Client
	}{
		{"1 Node all roles " + config.StandardClientName.String(), nodeRolesAll, r.standardUserClient},
		{"2 nodes - etcd/cp roles per 1 node " + config.StandardClientName.String(), nodeRolesShared, r.standardUserClient},
		{"3 nodes - 1 role per node " + config.StandardClientName.String(), nodeRolesDedicated, r.standardUserClient},
	}

	for _, tt := range tests {
		clusterConfig := *r.clusterConfig
		clusterConfig.Nodepools = tt.nodeRoles

		clusterName := namegen.AppendRandomString(provisioning.TFP)

		r.Run((tt.name), func() {
			provisioning.Provision(r.T(), clusterName, r.terraformConfig, &clusterConfig, r.terraformOptions)
			provisioning.VerifyCluster(r.T(), tt.client, clusterName, r.terraformConfig, r.terraformOptions, &clusterConfig)
			cleanup.Cleanup(r.T(), r.terraformOptions)
		})
	}
}

func (r *ProvisionTestSuite) TestProvisionDynamicInput() {
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
			cleanup.Cleanup(r.T(), r.terraformOptions)
		})
	}
}

func TestProvisionTestSuite(t *testing.T) {
	suite.Run(t, new(ProvisionTestSuite))
}
