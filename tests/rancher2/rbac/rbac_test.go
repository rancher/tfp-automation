//go:build validation || recurring

package rbac

import (
	"testing"

	clusterActions "github.com/rancher/tests/actions/clusters"
	provisioningActions "github.com/rancher/tests/actions/provisioning"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	nested "github.com/rancher/tfp-automation/tests/extensions/nestedModules"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	rb "github.com/rancher/tfp-automation/tests/extensions/rbac"
	"github.com/rancher/tfp-automation/tests/rancher2/resources"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestTfpRBAC(t *testing.T) {
	t.Parallel()

	r := resources.Setup(t)

	tests := []struct {
		name     string
		module   string
		rbacRole config.Role
	}{
		{"RKE2_Cluster_Owner", r.RKE2Module, config.ClusterOwner},
		{"RKE2_Project_Owner", r.RKE2Module, config.ProjectOwner},
		{"K3S_Cluster_Owner", r.K3SModule, config.ClusterOwner},
		{"K3S_Project_Owner", r.K3SModule, config.ProjectOwner},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(r.CattleConfig)
			rancher.AdminToken = r.StandardToken
			terraform.Module = tt.module
			terratest.Nodepools = r.NodeRoles

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(r.TerraformConfig, r.TerratestConfig, r.TerraformOptions, tt.name, "/modules/rancher2")
			require.NoError(t, err)
			defer func() {
				resources.RemoveDirectory(t, r.RancherConfig, nestedRancherModuleDir)
			}()

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()

			terratest, err = provisioning.GetK8sVersion(r.Client, terraform, terratest)
			require.NoError(t, err)

			terraform = provisioning.UniquifyTerraform(terraform)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.TerratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(t, perTestTerraformOptions, keyPath)

			logrus.Infof("Provisioning cluster (%s)", terraform.ResourcePrefix)
			clusters, _ := provisioning.Provision(t, r.Client, r.StandardUserClient, rancher, terraform, terratest, perTestTerraformOptions, newFile, rootBody, file, false, false, false, "", nestedRancherModuleDir)

			logrus.Infof("Verifying the cluster is ready (%s)", clusters[0].Name)
			err = provisioningActions.VerifyClusterReady(r.Client, clusters[0])
			require.NoError(t, err)

			logrus.Infof("Verifying service account token secret (%s)", clusters[0].Name)
			err = clusterActions.VerifyServiceAccountTokenSecret(r.Client, clusters[0].Name)
			require.NoError(t, err)

			logrus.Infof("Verifying cluster pods (%s)", clusters[0].Name)
			err = pods.VerifyClusterPods(r.Client, clusters[0])
			require.NoError(t, err)

			rb.RBAC(t, r.Client, rancher, terraform, terratest, perTestTerraformOptions, []map[string]any{r.CattleConfig}, tt.rbacRole, newFile, rootBody, file, nestedRancherModuleDir)

			params := tfpQase.GetProvisioningSchemaParams(r.TerraformConfig, r.TerratestConfig)
			err = qase.UpdateSchemaParameters(tt.name, params)
			if err != nil {
				logrus.Warningf("Failed to upload schema parameters %s", err)
			}
		})
	}

	if r.TerratestConfig.LocalQaseReporting {
		results.ReportTest(r.TerratestConfig)
	}
}
