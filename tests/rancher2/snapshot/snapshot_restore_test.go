//go:build validation || recurring

package snapshot

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	clusterActions "github.com/rancher/tests/actions/clusters"
	provisioningActions "github.com/rancher/tests/actions/provisioning"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	nested "github.com/rancher/tfp-automation/tests/extensions/nestedModules"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/infrastructure/ranchers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SnapshotRestoreTestSuite struct {
	suite.Suite
	client             *rancher.Client
	standardUserClient *rancher.Client
	session            *session.Session
	cattleConfig       map[string]any
	rancherConfig      *rancher.Config
	terraformConfig    *config.TerraformConfig
	terratestConfig    *config.TerratestConfig
	terraformOptions   *terraform.Options
}

func (s *SnapshotRestoreTestSuite) SetupSuite() {
	s.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	s.rancherConfig, s.terraformConfig, s.terratestConfig, _ = config.LoadTFPConfigs(s.cattleConfig)

	testSession := session.NewSession()
	s.session = testSession

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, s.terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(s.T(), s.terraformConfig, s.terratestConfig, keyPath)

	s.terraformOptions = terraformOptions

	client, err := ranchers.PostRancherSetup(s.T(), s.terraformOptions, s.rancherConfig, s.session, s.rancherConfig.Host, keyPath, false)
	require.NoError(s.T(), err)

	s.client = client
}

func (s *SnapshotRestoreTestSuite) TestTfpSnapshotRestore() {
	var err error
	var testUser, testPassword string
	var clusterIDs []string

	s.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(s.client)
	require.NoError(s.T(), err)

	standardUserToken, err := ranchers.CreateStandardUserToken(s.T(), s.terraformOptions, s.rancherConfig, testUser, testPassword)
	require.NoError(s.T(), err)

	standardToken := standardUserToken.Token

	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}
	rke2Module, _, _, k3sModule := provisioning.DownstreamClusterModules(s.terraformConfig)

	snapshotRestoreNone := config.TerratestConfig{
		SnapshotInput: config.Snapshots{
			SnapshotRestore: "none",
		},
	}

	tests := []struct {
		name         string
		module       string
		nodeRoles    []config.Nodepool
		etcdSnapshot config.TerratestConfig
	}{
		{"RKE2_Snapshot_Restore", rke2Module, nodeRolesDedicated, snapshotRestoreNone},
		{"K3S_Snapshot_Restore", k3sModule, nodeRolesDedicated, snapshotRestoreNone},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(s.terraformConfig, s.terratestConfig, s.terraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()

			configMap, err := provisioning.UniquifyTerraform([]map[string]any{s.cattleConfig})
			require.NoError(t, err)

			_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, configMap[0])
			require.NoError(s.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
			require.NoError(s.T(), err)

			_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
			require.NoError(s.T(), err)

			_, err = operations.ReplaceValue([]string{"terratest", "snapshotInput", "snapshotRestore"}, tt.etcdSnapshot.SnapshotInput.SnapshotRestore, configMap[0])
			require.NoError(s.T(), err)

			provisioning.GetK8sVersion(s.T(), s.client, s.terratestConfig, s.terraformConfig, configMap)

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, s.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(s.T(), perTestTerraformOptions, keyPath)

			clusters, _ := provisioning.Provision(s.T(), s.client, s.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, configMap, newFile, rootBody, file, false, false, false, clusterIDs, nil, nestedRancherModuleDir)
			err = provisioningActions.VerifyClusterReady(s.client, clusters[0])
			require.NoError(s.T(), err)
			err = clusterActions.VerifyServiceAccountTokenSecret(s.client, clusters[0].Name)
			require.NoError(s.T(), err)

			cluster, err := s.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraform.ResourcePrefix)
			require.NoError(s.T(), err)

			err = pods.VerifyClusterPods(s.client, cluster)
			require.NoError(s.T(), err)

			RestoreSnapshot(s.T(), s.client, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, configMap, newFile, rootBody, file, nestedRancherModuleDir)

			params := tfpQase.GetProvisioningSchemaParams(configMap[0])
			err = qase.UpdateSchemaParameters(tt.name, params)
			if err != nil {
				logrus.Warningf("Failed to upload schema parameters %s", err)
			}
		})
	}

	if s.terratestConfig.LocalQaseReporting {
		results.ReportTest(s.terratestConfig)
	}
}

func TestTfpSnapshotRestoreTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotRestoreTestSuite))
}
