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
	"github.com/rancher/tfp-automation/defaults/clustertypes"
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
	s.rancherConfig, s.terraformConfig, s.terratestConfig = config.LoadTFPConfigs(s.cattleConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, s.terratestConfig.PathToRepo, "")
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

	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(s.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{s.cattleConfig})
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		require.NoError(s.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "snapshotInput", "snapshotRestore"}, tt.etcdSnapshot.SnapshotInput.SnapshotRestore, configMap[0])
		require.NoError(s.T(), err)

		provisioning.GetK8sVersion(s.T(), s.client, s.terratestConfig, s.terraformConfig, configs.DefaultK8sVersion, configMap)

		rancher, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		tt.name = tt.name + " Module: " + s.terraformConfig.Module + " Kubernetes version: " + terratest.KubernetesVersion

		if strings.Contains(s.terraformConfig.Module, clustertypes.RKE1) {
			s.T().Skip("RKE1 is not supported")
		}

		s.Run(tt.name, func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, s.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(s.T(), s.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(s.T(), s.client)
			require.NoError(s.T(), err)

			clusterIDs, _ := provisioning.Provision(s.T(), s.client, rancher, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file, false, false, false, nil)
			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)

			RestoreSnapshot(s.T(), s.client, rancher, terraform, terratest, testUser, testPassword, s.terraformOptions, configMap, newFile, rootBody, file)
			provisioning.VerifyClustersState(s.T(), adminClient, clusterIDs)
		})
	}

	if s.terratestConfig.LocalQaseReporting {
		qase.ReportTest(s.terratestConfig)
	}
}

func TestTfpSnapshotRestoreTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotRestoreTestSuite))
}
