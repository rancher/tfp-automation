package snapshot

import (
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SnapshotRestoreTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	cattleConfig     map[string]any
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

	s.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	configMap, err := provisioning.UniquifyTerraform([]map[string]any{s.cattleConfig})
	require.NoError(s.T(), err)

	s.cattleConfig = configMap[0]
	s.rancherConfig, s.terraformConfig, s.terratestConfig = config.LoadTFPConfigs(s.cattleConfig)

	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
	terraformOptions := framework.Setup(s.T(), s.terraformConfig, s.terratestConfig, keyPath)
	s.terraformOptions = terraformOptions
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
		configMap := []map[string]any{s.cattleConfig}

		operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		operations.ReplaceValue([]string{"terratest", "snapshotInput", "snapshotRestore"}, tt.etcdSnapshot.SnapshotInput.SnapshotRestore, configMap[0])

		provisioning.GetK8sVersion(s.T(), s.client, s.terratestConfig, s.terraformConfig, configs.DefaultK8sVersion, configMap)

		_, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		tt.name = tt.name + " Module: " + s.terraformConfig.Module + " Kubernetes version: " + terratest.KubernetesVersion

		if strings.Contains(s.terraformConfig.Module, "rke1") {
			s.T().Skip("RKE1 is not supported")
		}

		testUser, testPassword := configs.CreateTestCredentials()

		s.Run(tt.name, func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
			defer cleanup.Cleanup(s.T(), s.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(s.T(), s.client)
			require.NoError(s.T(), err)

			configMap := []map[string]any{s.cattleConfig}

			clusterIDs := provisioning.Provision(s.T(), s.client, s.rancherConfig, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, false)
			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)

			snapshotRestore(s.T(), s.client, s.rancherConfig, s.terraformConfig, terratest, testUser, testPassword, s.terraformOptions, configMap)
			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)
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
		configMap := []map[string]any{s.cattleConfig}
		provisioning.GetK8sVersion(s.T(), s.client, s.terratestConfig, s.terraformConfig, configs.DefaultK8sVersion, configMap)

		tt.name = tt.name + " Module: " + s.terraformConfig.Module + " Kubernetes version: " + s.terratestConfig.KubernetesVersion

		testUser, testPassword := configs.CreateTestCredentials()

		s.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
			defer cleanup.Cleanup(s.T(), s.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(s.T(), s.client)
			require.NoError(s.T(), err)

			clusterIDs := provisioning.Provision(s.T(), s.client, s.rancherConfig, s.terraformConfig, s.terratestConfig, testUser, testPassword, s.terraformOptions, nil, false)
			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)

			snapshotRestore(s.T(), s.client, s.rancherConfig, s.terraformConfig, s.terratestConfig, testUser, testPassword, s.terraformOptions, configMap)
			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)
		})
	}

	if s.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpSnapshotRestoreTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotRestoreTestSuite))
}
