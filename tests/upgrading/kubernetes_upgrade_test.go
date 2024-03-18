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

type KubernetesUpgradeTestSuite struct {
	suite.Suite
	client             *rancher.Client
	standardUserClient *rancher.Client
	session            *session.Session
	terraformConfig    *config.TerraformConfig
	clusterConfig      *config.TerratestConfig
	terraformOptions   *terraform.Options
}

func (k *KubernetesUpgradeTestSuite) SetupSuite() {
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

func (k *KubernetesUpgradeTestSuite) TestTfpKubernetesUpgrade() {
	nodeRolesAll := []config.Nodepool{config.AllRolesNodePool}
	nodeRolesShared := []config.Nodepool{config.EtcdControlPlaneNodePool, config.WorkerNodePool}
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		client    *rancher.Client
	}{
		{"1 Node all roles " + config.StandardClientName.String(), nodeRolesAll, k.standardUserClient},
		{"2 nodes - etcd/cp roles per 1 node " + config.StandardClientName.String(), nodeRolesShared, k.standardUserClient},
		{"3 nodes - 1 role per node " + config.StandardClientName.String(), nodeRolesDedicated, k.standardUserClient},
	}

	for _, tt := range tests {
		clusterConfig := *k.clusterConfig
		clusterConfig.Nodepools = tt.nodeRoles

		tt.name = tt.name + " Module: " + k.terraformConfig.Module + " Kubernetes version: " + k.clusterConfig.KubernetesVersion

		clusterName := namegen.AppendRandomString(provisioning.TFP)

		k.Run((tt.name), func() {
			defer cleanup.Cleanup(k.T(), k.terraformOptions)

			provisioning.Provision(k.T(), tt.client, clusterName, &clusterConfig, k.terraformOptions)
			provisioning.VerifyCluster(k.T(), tt.client, clusterName, k.terraformConfig, k.terraformOptions, &clusterConfig)

			provisioning.KubernetesUpgrade(k.T(), tt.client, clusterName, k.terraformOptions, k.terraformConfig, &clusterConfig)
			provisioning.VerifyCluster(k.T(), tt.client, clusterName, k.terraformConfig, k.terraformOptions, &clusterConfig)
			provisioning.VerifyUpgradedKubernetesVersion(k.T(), tt.client, k.terraformConfig, clusterName,
				k.clusterConfig.UpgradedKubernetesVersion)
		})
	}
}

func (k *KubernetesUpgradeTestSuite) TestTfpKubernetesUpgradeDynamicInput() {
	tests := []struct {
		name   string
		client *rancher.Client
	}{
		{config.StandardClientName.String(), k.standardUserClient},
	}

	for _, tt := range tests {
		tt.name = tt.name + " Module: " + k.terraformConfig.Module + " Kubernetes version: " + k.clusterConfig.KubernetesVersion

		clusterName := namegen.AppendRandomString(provisioning.TFP)

		k.Run((tt.name), func() {
			defer cleanup.Cleanup(k.T(), k.terraformOptions)

			provisioning.Provision(k.T(), tt.client, clusterName, k.clusterConfig, k.terraformOptions)
			provisioning.VerifyCluster(k.T(), k.client, clusterName, k.terraformConfig, k.terraformOptions, k.clusterConfig)

			provisioning.KubernetesUpgrade(k.T(), tt.client, clusterName, k.terraformOptions, k.terraformConfig, k.clusterConfig)
			provisioning.VerifyCluster(k.T(), k.client, clusterName, k.terraformConfig, k.terraformOptions, k.clusterConfig)
			provisioning.VerifyUpgradedKubernetesVersion(k.T(), k.client, k.terraformConfig, clusterName,
				k.clusterConfig.UpgradedKubernetesVersion)
		})
	}
}

func TestTfpKubernetesUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(KubernetesUpgradeTestSuite))
}
