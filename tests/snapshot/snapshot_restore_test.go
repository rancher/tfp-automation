package snapshot

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/framework"
	cleanup "github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SnapshotRestoreTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	terraformConfig  *config.TerraformConfig
	clusterConfig    *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (s *SnapshotRestoreTestSuite) SetupSuite() {
	testSession := session.NewSession()
	s.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(s.T(), err)

	s.client = client

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, terraformConfig)

	s.terraformConfig = terraformConfig

	clusterConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, clusterConfig)

	s.clusterConfig = clusterConfig

	terraformOptions := framework.Setup(s.T())
	s.terraformOptions = terraformOptions

	provisioning.GetK8sVersion(s.T(), s.client, s.clusterConfig, s.terraformConfig, configs.DefaultK8sVersion)
}

func (s *SnapshotRestoreTestSuite) TestTfpSnapshotRestoreETCDOnly() {
	nodeRolesAll := []config.Nodepool{config.AllRolesNodePool}
	nodeRolesShared := []config.Nodepool{config.EtcdControlPlaneNodePool, config.WorkerNodePool}
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	snapshotRestoreNone := config.TerratestConfig{
		SnapshotInput: config.Snapshots{
			UpgradeKubernetesVersion: "",
			SnapshotRestore:          "none",
		},
	}

	tests := []struct {
		name         string
		nodeRoles    []config.Nodepool
		etcdSnapshot config.TerratestConfig
	}{
		{"Restore etcd only: 1 node all roles", nodeRolesAll, snapshotRestoreNone},
		{"Restore etcd only: 2 nodes shared roles", nodeRolesShared, snapshotRestoreNone},
		{"Restore etcd only: 3 nodes dedicated roles", nodeRolesDedicated, snapshotRestoreNone},
	}

	for _, tt := range tests {
		clusterConfig := *s.clusterConfig
		clusterConfig.Nodepools = tt.nodeRoles
		clusterConfig.SnapshotInput.UpgradeKubernetesVersion = tt.etcdSnapshot.SnapshotInput.UpgradeKubernetesVersion
		clusterConfig.SnapshotInput.SnapshotRestore = tt.etcdSnapshot.SnapshotInput.SnapshotRestore

		tt.name = tt.name + " Module: " + s.terraformConfig.Module + " Kubernetes version: " + s.clusterConfig.KubernetesVersion

		clusterName := namegen.AppendRandomString(configs.TFP)
		poolName := namegen.AppendRandomString(configs.TFP)

		s.Run(tt.name, func() {
			defer cleanup.Cleanup(s.T(), s.terraformOptions)

			provisioning.Provision(s.T(), clusterName, poolName, &clusterConfig, s.terraformOptions)
			provisioning.VerifyCluster(s.T(), s.client, clusterName, s.terraformConfig, s.terraformOptions, &clusterConfig)

			snapshotRestore(s.T(), s.client, clusterName, poolName, &clusterConfig, s.terraformOptions)
		})
	}
}

func (s *SnapshotRestoreTestSuite) TestTfpSnapshotRestoreETCDOnlyDynamicInput() {
	if s.clusterConfig.SnapshotInput == (config.Snapshots{}) {
		s.T().Skip()
	}

	tests := []struct {
		name string
	}{
		{config.StandardClientName.String()},
	}

	for _, tt := range tests {
		tt.name = tt.name + " Module: " + s.terraformConfig.Module + " Kubernetes version: " + s.clusterConfig.KubernetesVersion

		clusterName := namegen.AppendRandomString(configs.TFP)
		poolName := namegen.AppendRandomString(configs.TFP)

		s.Run((tt.name), func() {
			defer cleanup.Cleanup(s.T(), s.terraformOptions)

			provisioning.Provision(s.T(), clusterName, poolName, s.clusterConfig, s.terraformOptions)
			provisioning.VerifyCluster(s.T(), s.client, clusterName, s.terraformConfig, s.terraformOptions, s.clusterConfig)

			snapshotRestore(s.T(), s.client, clusterName, poolName, s.clusterConfig, s.terraformOptions)
		})
	}
}

func TestTfpSnapshotRestoreTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotRestoreTestSuite))
}
