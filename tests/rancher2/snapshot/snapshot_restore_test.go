package snapshot

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/framework"
	cleanup "github.com/rancher/tfp-automation/framework/cleanup/rancher2"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SnapshotRestoreTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	rancherConfig    *rancher.Config
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (s *SnapshotRestoreTestSuite) SetupSuite() {
	testSession := session.NewSession()
	s.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(s.T(), err)

	s.client = client

	rancherConfig := new(rancher.Config)
	ranchFrame.LoadConfig(configs.Rancher, rancherConfig)

	s.rancherConfig = rancherConfig

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, terraformConfig)

	s.terraformConfig = terraformConfig

	terratestConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, terratestConfig)

	s.terratestConfig = terratestConfig

	terraformOptions := framework.Rancher2Setup(s.T(), s.rancherConfig, s.terraformConfig, s.terratestConfig)
	s.terraformOptions = terraformOptions

	provisioning.GetK8sVersion(s.T(), s.client, s.terratestConfig, s.terraformConfig, configs.DefaultK8sVersion)
}

func (s *SnapshotRestoreTestSuite) TestTfpSnapshotRestore() {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	snapshotRestoreNone := config.TerratestConfig{
		SnapshotInput: config.Snapshots{
			SnapshotRestore: "none",
		},
	}

	tests := []struct {
		name         string
		nodeRoles    []config.Nodepool
		etcdSnapshot config.TerratestConfig
	}{
		{"Restore etcd only", nodeRolesDedicated, snapshotRestoreNone},
	}

	for _, tt := range tests {
		terratestConfig := *s.terratestConfig
		terratestConfig.Nodepools = tt.nodeRoles
		terratestConfig.SnapshotInput.SnapshotRestore = tt.etcdSnapshot.SnapshotInput.SnapshotRestore

		tt.name = tt.name + " Module: " + s.terraformConfig.Module + " Kubernetes version: " + s.terratestConfig.KubernetesVersion

		if strings.Contains(s.terraformConfig.Module, "rke1") {
			s.T().Skip("RKE1 is not supported")
		}

		testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

		s.Run(tt.name, func() {
			defer cleanup.ConfigCleanup(s.T(), s.terraformOptions)

			adminClient, err := provisioning.FetchAdminClient(s.T(), s.client)
			require.NoError(s.T(), err)

			provisioning.Provision(s.T(), s.client, s.rancherConfig, s.terraformConfig, &terratestConfig, testUser, testPassword, clusterName, poolName, s.terraformOptions, nil)
			provisioning.VerifyCluster(s.T(), adminClient, clusterName, s.terraformConfig, s.terraformOptions, &terratestConfig)

			snapshotRestore(s.T(), s.client, s.rancherConfig, s.terraformConfig, &terratestConfig, testUser, testPassword, clusterName, poolName, s.terraformOptions)
		})
	}

	if s.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func (s *SnapshotRestoreTestSuite) TestTfpSnapshotRestoreDynamicInput() {
	if s.terratestConfig.SnapshotInput == (config.Snapshots{}) {
		s.T().Skip()
	}

	tests := []struct {
		name string
	}{
		{config.StandardClientName.String()},
	}

	for _, tt := range tests {
		tt.name = tt.name + " Module: " + s.terraformConfig.Module + " Kubernetes version: " + s.terratestConfig.KubernetesVersion

		testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

		s.Run((tt.name), func() {
			defer cleanup.ConfigCleanup(s.T(), s.terraformOptions)

			adminClient, err := provisioning.FetchAdminClient(s.T(), s.client)
			require.NoError(s.T(), err)

			provisioning.Provision(s.T(), s.client, s.rancherConfig, s.terraformConfig, s.terratestConfig, testUser, testPassword, clusterName, poolName, s.terraformOptions, nil)
			provisioning.VerifyCluster(s.T(), adminClient, clusterName, s.terraformConfig, s.terraformOptions, s.terratestConfig)

			snapshotRestore(s.T(), s.client, s.rancherConfig, s.terraformConfig, s.terratestConfig, testUser, testPassword, clusterName, poolName, s.terraformOptions)
		})
	}

	if s.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpSnapshotRestoreTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotRestoreTestSuite))
}
