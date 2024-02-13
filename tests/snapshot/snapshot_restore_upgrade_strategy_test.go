package snapshot

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

type SnapshotRestoreUpgradeStrategyTestSuite struct {
	suite.Suite
	client             *rancher.Client
	standardUserClient *rancher.Client
	session            *session.Session
	terraformConfig    *config.TerraformConfig
	clusterConfig      *config.TerratestConfig
	terraformOptions   *terraform.Options
}

func (s *SnapshotRestoreUpgradeStrategyTestSuite) SetupSuite() {
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
}

func (s *SnapshotRestoreUpgradeStrategyTestSuite) TestSnapshotRestoreUpgradeStrategy() {
	nodeRolesAll := []config.Nodepool{config.AllRolesNodePool}
	nodeRolesShared := []config.Nodepool{config.EtcdControlPlaneNodePool, config.WorkerNodePool}
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	snapshotRestoreK8sVersion := config.TerratestConfig{
		SnapshotInput: config.Snapshots{
			UpgradeKubernetesVersion:     "",
			SnapshotRestore:              "kubernetesVersion",
			ControlPlaneConcurrencyValue: "15%",
			ControlPlaneUnavailableValue: "3",
			WorkerConcurrencyValue:       "20%",
			WorkerUnavailableValue:       "15%",
			RecurringRestores:            1,
		},
	}

	snapshotRestoreAll := config.TerratestConfig{
		SnapshotInput: config.Snapshots{
			UpgradeKubernetesVersion:     "",
			SnapshotRestore:              "all",
			ControlPlaneConcurrencyValue: "15%",
			ControlPlaneUnavailableValue: "3",
			WorkerConcurrencyValue:       "20%",
			WorkerUnavailableValue:       "15%",
			RecurringRestores:            1,
		},
	}

	tests := []struct {
		name         string
		nodeRoles    []config.Nodepool
		etcdSnapshot config.TerratestConfig
		client       *rancher.Client
	}{
		{"Restore Kubernetes version and etcd: 1 node all roles", nodeRolesAll, snapshotRestoreK8sVersion, s.client},
		{"Restore cluster config, Kubernetes version and etcd: 1 node all roles", nodeRolesAll, snapshotRestoreAll, s.client},
		{"Restore Kubernetes version and etcd: 2 nodes shared roles", nodeRolesShared, snapshotRestoreK8sVersion, s.client},
		{"Restore cluster config, Kubernetes version and etcd: 2 nodes shared roles", nodeRolesShared, snapshotRestoreAll, s.client},
		{"Restore Kubernetes version and etcd: 3 nodes dedicated roles", nodeRolesDedicated, snapshotRestoreK8sVersion, s.client},
		{"Restore cluster config, Kubernetes version and etcd: 3 nodes dedicated roles", nodeRolesDedicated, snapshotRestoreAll, s.client},
	}

	for _, tt := range tests {
		clusterConfig := *s.clusterConfig
		clusterConfig.Nodepools = tt.nodeRoles
		clusterConfig.SnapshotInput.UpgradeKubernetesVersion = tt.etcdSnapshot.SnapshotInput.UpgradeKubernetesVersion
		clusterConfig.SnapshotInput.SnapshotRestore = tt.etcdSnapshot.SnapshotInput.SnapshotRestore
		clusterConfig.SnapshotInput.ControlPlaneConcurrencyValue = tt.etcdSnapshot.SnapshotInput.ControlPlaneConcurrencyValue
		clusterConfig.SnapshotInput.ControlPlaneUnavailableValue = tt.etcdSnapshot.SnapshotInput.ControlPlaneUnavailableValue
		clusterConfig.SnapshotInput.WorkerConcurrencyValue = tt.etcdSnapshot.SnapshotInput.WorkerConcurrencyValue
		clusterConfig.SnapshotInput.WorkerUnavailableValue = tt.etcdSnapshot.SnapshotInput.WorkerUnavailableValue
		clusterConfig.SnapshotInput.RecurringRestores = tt.etcdSnapshot.SnapshotInput.RecurringRestores

		clusterName := namegen.AppendRandomString(provisioning.TFP)

		s.Run(tt.name, func() {
			provisioning.Provision(s.T(), clusterName, s.terraformConfig, &clusterConfig, s.terraformOptions)
			provisioning.VerifyCluster(s.T(), tt.client, clusterName, s.terraformConfig, s.terraformOptions, &clusterConfig)

			snapshotRestore(s.T(), s.client, clusterName, &clusterConfig, s.terraformOptions)

			cleanup.Cleanup(s.T(), s.terraformOptions)
		})
	}
}

func (s *SnapshotRestoreUpgradeStrategyTestSuite) TestSnapshotRestoreUpgradeStrategyDynamicInput() {
	if s.clusterConfig.SnapshotInput == (config.Snapshots{}) {
		s.T().Skip()
	}

	tests := []struct {
		name   string
		client *rancher.Client
	}{
		{config.StandardClientName.String(), s.standardUserClient},
	}

	for _, tt := range tests {
		clusterName := namegen.AppendRandomString(provisioning.TFP)

		s.Run((tt.name), func() {
			provisioning.Provision(s.T(), clusterName, s.terraformConfig, s.clusterConfig, s.terraformOptions)
			provisioning.VerifyCluster(s.T(), s.client, clusterName, s.terraformConfig, s.terraformOptions, s.clusterConfig)

			snapshotRestore(s.T(), s.client, clusterName, s.clusterConfig, s.terraformOptions)

			cleanup.Cleanup(s.T(), s.terraformOptions)
		})
	}
}

func TestSnapshotRestoreUpgradeStrategyTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotRestoreUpgradeStrategyTestSuite))
}
