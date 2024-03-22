package psact

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

type PSACTTestSuite struct {
	suite.Suite
	client             *rancher.Client
	standardUserClient *rancher.Client
	session            *session.Session
	terraformConfig    *config.TerraformConfig
	clusterConfig      *config.TerratestConfig
	terraformOptions   *terraform.Options
}

func (p *PSACTTestSuite) SetupSuite() {
	testSession := session.NewSession()
	p.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(p.T(), err)

	p.client = client

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
	require.NoError(p.T(), err)

	newUser.Password = user.Password

	standardUserClient, err := client.AsUser(newUser)
	require.NoError(p.T(), err)

	p.standardUserClient = standardUserClient

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

func (p *PSACTTestSuite) TestTfpPSACT() {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		psact     config.PSACT
		client    *rancher.Client
	}{
		{
			name:      "Rancher Privileged " + config.StandardClientName.String(),
			nodeRoles: nodeRolesDedicated,
			psact:     "rancher-privileged",
			client:    p.standardUserClient,
		},
		{
			name:      "Rancher Restricted " + config.StandardClientName.String(),
			nodeRoles: nodeRolesDedicated,
			psact:     "rancher-restricted",
			client:    p.standardUserClient,
		},
		{
			name:      "Rancher Baseline " + config.AdminClientName.String(),
			nodeRoles: nodeRolesDedicated,
			psact:     "rancher-baseline",
			client:    p.client,
		},
	}

	for _, tt := range tests {
		clusterConfig := *p.clusterConfig
		clusterConfig.Nodepools = tt.nodeRoles
		clusterConfig.PSACT = string(tt.psact)

		tt.name = tt.name + " Module: " + p.terraformConfig.Module + " Kubernetes version: " + p.clusterConfig.KubernetesVersion

		clusterName := namegen.AppendRandomString(provisioning.TFP)
		poolName := namegen.AppendRandomString(provisioning.TFP)

		p.Run((tt.name), func() {
			defer cleanup.Cleanup(p.T(), p.terraformOptions)

			provisioning.Provision(p.T(), tt.client, clusterName, poolName, &clusterConfig, p.terraformOptions)
			provisioning.VerifyCluster(p.T(), tt.client, clusterName, p.terraformConfig, p.terraformOptions, &clusterConfig)
		})
	}
}

func TestTfpPSACTTestSuite(t *testing.T) {
	suite.Run(t, new(PSACTTestSuite))
}
