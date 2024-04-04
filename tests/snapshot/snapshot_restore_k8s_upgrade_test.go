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
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SnapshotRestoreK8sUpgradeTestSuite struct {
	suite.Suite
	client             *rancher.Client
	standardUserClient *rancher.Client
	session            *session.Session
	terraformConfig    *config.TerraformConfig
	clusterConfig      *config.TerratestConfig
	terraformOptions   *terraform.Options
}

func (s *SnapshotRestoreK8sUpgradeTestSuite) SetupSuite() {
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

func (s *SnapshotRestoreK8sUpgradeTestSuite) TestTfpSnapshotRestoreK8sUpgrade() {
	nodeRolesAll := []config.Nodepool{config.AllRolesNodePool}
	nodeRolesShared := []config.Nodepool{config.EtcdControlPlaneNodePool, config.WorkerNodePool}
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	snapshotRestoreK8sVersion := config.TerratestConfig{
		SnapshotInput: config.Snapshots{
			UpgradeKubernetesVersion: "",
			SnapshotRestore:          "kubernetesVersion",
		},
	}

	snapshotRestoreAll := config.TerratestConfig{
		SnapshotInput: config.Snapshots{
			UpgradeKubernetesVersion: "",
			SnapshotRestore:          "all",
		},
	}

	tests := []struct {
		name         string
		nodeRoles    []config.Nodepool
		etcdSnapshot config.TerratestConfig
		client       *rancher.Client
	}{
		{"Restore K8s version and etcd: all roles", nodeRolesAll, snapshotRestoreK8sVersion, s.standardUserClient},
		{"Restore cluster config, K8s version and etcd: all roles", nodeRolesAll, snapshotRestoreAll, s.standardUserClient},
		{"Restore K8s version and etcd: shared roles", nodeRolesShared, snapshotRestoreK8sVersion, s.standardUserClient},
		{"Restore cluster config, K8s version and etcd: shared roles", nodeRolesShared, snapshotRestoreAll, s.standardUserClient},
		{"Restore K8s version and etcd: dedicated roles", nodeRolesDedicated, snapshotRestoreK8sVersion, s.standardUserClient},
		{"Restore cluster config, K8s version and etcd: dedicated roles", nodeRolesDedicated, snapshotRestoreAll, s.standardUserClient},
	}

	for _, tt := range tests {
		clusterConfig := *s.clusterConfig
		clusterConfig.Nodepools = tt.nodeRoles
		clusterConfig.SnapshotInput.UpgradeKubernetesVersion = tt.etcdSnapshot.SnapshotInput.UpgradeKubernetesVersion
		clusterConfig.SnapshotInput.SnapshotRestore = tt.etcdSnapshot.SnapshotInput.SnapshotRestore

		clusterName := namegen.AppendRandomString(provisioning.TFP)
		poolName := namegen.AppendRandomString(provisioning.TFP)

		s.Run(tt.name, func() {
			defer cleanup.Cleanup(s.T(), s.terraformOptions)

			logrus.Infof("Module: %s", s.terraformConfig.Module)
			logrus.Infof("Kubernetes version: %s", s.clusterConfig.KubernetesVersion)

			provisioning.Provision(s.T(), tt.client, clusterName, poolName, &clusterConfig, s.terraformOptions)
			provisioning.VerifyCluster(s.T(), tt.client, clusterName, s.terraformConfig, s.terraformOptions, &clusterConfig)

			snapshotRestore(s.T(), s.client, clusterName, poolName, &clusterConfig, s.terraformOptions)
		})
	}
}

func (s *SnapshotRestoreK8sUpgradeTestSuite) TestTfpSnapshotRestoreK8sUpgradeDynamicInput() {
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
		poolName := namegen.AppendRandomString(provisioning.TFP)

		s.Run((tt.name), func() {
			defer cleanup.Cleanup(s.T(), s.terraformOptions)

			logrus.Infof("Module: %s", s.terraformConfig.Module)
			logrus.Infof("Kubernetes version: %s", s.clusterConfig.KubernetesVersion)

			provisioning.Provision(s.T(), tt.client, clusterName, poolName, s.clusterConfig, s.terraformOptions)
			provisioning.VerifyCluster(s.T(), s.client, clusterName, s.terraformConfig, s.terraformOptions, s.clusterConfig)

			snapshotRestore(s.T(), s.client, clusterName, poolName, s.clusterConfig, s.terraformOptions)
		})
	}
}

func TestTfpSnapshotRestoreK8sUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotRestoreK8sUpgradeTestSuite))
}
