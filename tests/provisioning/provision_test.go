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
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework"
	cleanup "github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/sirupsen/logrus"
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

func (p *ProvisionTestSuite) SetupSuite() {
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

func (p *ProvisionTestSuite) TestTfpProvision() {
	nodeRolesAll := []config.Nodepool{config.AllRolesNodePool}
	nodeRolesShared := []config.Nodepool{config.EtcdControlPlaneNodePool, config.WorkerNodePool}
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		client    *rancher.Client
	}{
		{"1 Node all roles " + config.StandardClientName.String(), nodeRolesAll, p.standardUserClient},
		{"2 nodes - etcd|cp roles per 1 node " + config.StandardClientName.String(), nodeRolesShared, p.standardUserClient},
		{"3 nodes - 1 role per node " + config.StandardClientName.String(), nodeRolesDedicated, p.standardUserClient},
	}

	for _, tt := range tests {
		clusterConfig := *p.clusterConfig
		clusterConfig.Nodepools = tt.nodeRoles

		clusterName := namegen.AppendRandomString(provisioning.TFP)
		poolName := namegen.AppendRandomString(provisioning.TFP)

		p.Run((tt.name), func() {
			defer cleanup.Cleanup(p.T(), p.terraformOptions)

			logrus.Infof("Module: %s", p.terraformConfig.Module)
			logrus.Infof("Kubernetes version: %s", p.clusterConfig.KubernetesVersion)

			provisioning.Provision(p.T(), tt.client, clusterName, poolName, &clusterConfig, p.terraformOptions)
			provisioning.VerifyCluster(p.T(), tt.client, clusterName, p.terraformConfig, p.terraformOptions, &clusterConfig)
		})
	}
}

func (p *ProvisionTestSuite) TestTfpProvisionDynamicInput() {
	tests := []struct {
		name   string
		client *rancher.Client
	}{
		{config.StandardClientName.String(), p.standardUserClient},
	}

	for _, tt := range tests {
		clusterName := namegen.AppendRandomString(provisioning.TFP)
		poolName := namegen.AppendRandomString(provisioning.TFP)

		p.Run((tt.name), func() {
			defer cleanup.Cleanup(p.T(), p.terraformOptions)

			logrus.Infof("Module: %s", p.terraformConfig.Module)
			logrus.Infof("Kubernetes version: %s", p.clusterConfig.KubernetesVersion)

			provisioning.Provision(p.T(), tt.client, clusterName, poolName, p.clusterConfig, p.terraformOptions)
			provisioning.VerifyCluster(p.T(), p.client, clusterName, p.terraformConfig, p.terraformOptions, p.clusterConfig)
		})
	}
}

func TestTfpProvisionTestSuite(t *testing.T) {
	suite.Run(t, new(ProvisionTestSuite))
}
