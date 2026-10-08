//go:build validation || recurring

package snapshot

import (
	"testing"

	clusterActions "github.com/rancher/tests/actions/clusters"
	provisioningActions "github.com/rancher/tests/actions/provisioning"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	nested "github.com/rancher/tfp-automation/tests/extensions/nestedModules"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/rancher2/resources"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestTfpSnapshotRestore(t *testing.T) {
	t.Parallel()

	s := resources.Setup(t)

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
		{"RKE2_Snapshot_Restore", s.RKE2Module, s.NodeRoles, snapshotRestoreNone},
		{"K3S_Snapshot_Restore", s.K3SModule, s.NodeRoles, snapshotRestoreNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(s.CattleConfig)
			rancher.AdminToken = s.StandardToken
			terraform.Module = tt.module
			terratest.Nodepools = tt.nodeRoles
			terratest.SnapshotInput.SnapshotRestore = tt.etcdSnapshot.SnapshotInput.SnapshotRestore

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(s.TerraformConfig, s.TerratestConfig, s.TerraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer func() {
				resources.RemoveDirectory(t, s.RancherConfig, nestedRancherModuleDir)
			}()

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()

			terratest, err = provisioning.GetK8sVersion(s.Client, terraform, terratest)
			require.NoError(t, err)

			terraform = provisioning.UniquifyTerraform(terraform)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, s.TerratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(t, perTestTerraformOptions, keyPath)

			logrus.Infof("Provisioning cluster (%s)", terraform.ResourcePrefix)
			clusters, _ := provisioning.Provision(t, s.Client, s.StandardUserClient, rancher, terraform, terratest, perTestTerraformOptions, newFile, rootBody, file, false, false, false, "", nestedRancherModuleDir)

			logrus.Infof("Verifying the cluster is ready (%s)", clusters[0].Name)
			err = provisioningActions.VerifyClusterReady(s.Client, clusters[0])
			require.NoError(t, err)

			logrus.Infof("Verifying service account token secret (%s)", clusters[0].Name)
			err = clusterActions.VerifyServiceAccountTokenSecret(s.Client, clusters[0].Name)
			require.NoError(t, err)

			logrus.Infof("Verifying cluster pods (%s)", clusters[0].Name)
			err = pods.VerifyClusterPods(s.Client, clusters[0])
			require.NoError(t, err)

			RestoreSnapshot(t, s.Client, rancher, terraform, terratest, perTestTerraformOptions, newFile, rootBody, file, nestedRancherModuleDir)

			params := tfpQase.GetProvisioningSchemaParams(s.TerraformConfig, s.TerratestConfig)
			err = qase.UpdateSchemaParameters(tt.name, params)
			if err != nil {
				logrus.Warningf("Failed to upload schema parameters %s", err)
			}
		})
	}

	if s.TerratestConfig.LocalQaseReporting {
		results.ReportTest(s.TerratestConfig)
	}
}
