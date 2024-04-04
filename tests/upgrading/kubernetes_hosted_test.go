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

type KubernetesUpgradeHostedTestSuite struct {
	suite.Suite
	client             *rancher.Client
	standardUserClient *rancher.Client
	session            *session.Session
	terraformConfig    *config.TerraformConfig
	clusterConfig      *config.TerratestConfig
	terraformOptions   *terraform.Options
}

func (k *KubernetesUpgradeHostedTestSuite) SetupSuite() {
	testSession := session.NewSession()
	k.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(k.T(), err)

	k.client = client

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
	require.NoError(k.T(), err)

	newUser.Password = user.Password

	standardUserClient, err := client.AsUser(newUser)
	require.NoError(k.T(), err)

	k.standardUserClient = standardUserClient

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, terraformConfig)

	k.terraformConfig = terraformConfig

	clusterConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, clusterConfig)

	k.clusterConfig = clusterConfig

	terraformOptions := framework.Setup(k.T())
	k.terraformOptions = terraformOptions
}

func (k *KubernetesUpgradeHostedTestSuite) TestTfpKubernetesUpgradeHosted() {
	tests := []struct {
		name   string
		client *rancher.Client
	}{
		{config.StandardClientName.String(), k.standardUserClient},
	}

	for _, tt := range tests {
		clusterName := namegen.AppendRandomString(provisioning.TFP)
		poolName := namegen.AppendRandomString(provisioning.TFP)

		k.Run((tt.name), func() {
			defer cleanup.Cleanup(k.T(), k.terraformOptions)

			logrus.Infof("Module: %s", k.terraformConfig.Module)
			logrus.Infof("Kubernetes version: %s", k.clusterConfig.KubernetesVersion)

			provisioning.Provision(k.T(), tt.client, clusterName, poolName, k.clusterConfig, k.terraformOptions)
			provisioning.VerifyCluster(k.T(), k.client, clusterName, k.terraformConfig, k.terraformOptions, k.clusterConfig)

			provisioning.KubernetesUpgrade(k.T(), tt.client, clusterName, poolName, k.terraformOptions, k.terraformConfig, k.clusterConfig)
			provisioning.VerifyCluster(k.T(), k.client, clusterName, k.terraformConfig, k.terraformOptions, k.clusterConfig)
			provisioning.VerifyUpgradedKubernetesVersion(k.T(), k.client, k.terraformConfig, clusterName,
				k.clusterConfig.UpgradedKubernetesVersion)
		})
	}
}

func TestTfpKubernetesUpgradeHostedTestSuite(t *testing.T) {
	suite.Run(t, new(KubernetesUpgradeHostedTestSuite))
}
