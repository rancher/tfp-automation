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
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ProvisionHostedTestSuite struct {
	suite.Suite
	client             *rancher.Client
	standardUserClient *rancher.Client
	session            *session.Session
	terraformConfig    *config.TerraformConfig
	clusterConfig      *config.TerratestConfig
	terraformOptions   *terraform.Options
}

func (p *ProvisionHostedTestSuite) SetupSuite() {
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

func (p *ProvisionHostedTestSuite) TestTfpProvisionHosted() {
	tests := []struct {
		name   string
		client *rancher.Client
	}{
		{config.StandardClientName.String(), p.standardUserClient},
	}

	for _, tt := range tests {
		tt.name = tt.name + " Module: " + p.terraformConfig.Module + " Kubernetes version: " + p.clusterConfig.KubernetesVersion

		clusterName := namegen.AppendRandomString(provisioning.TFP)
		poolName := namegen.AppendRandomString(provisioning.TFP)

		p.Run((tt.name), func() {
			defer cleanup.Cleanup(p.T(), p.terraformOptions)

			provisioning.Provision(p.T(), tt.client, clusterName, poolName, p.clusterConfig, p.terraformOptions)
			provisioning.VerifyCluster(p.T(), p.client, clusterName, p.terraformConfig, p.terraformOptions, p.clusterConfig)
		})
	}
}

func TestTfpProvisionHostedTestSuite(t *testing.T) {
	suite.Run(t, new(ProvisionHostedTestSuite))
}
