//go:build validation || recurring

package provisioning

import (
	"testing"

	clusterActions "github.com/rancher/tests/actions/clusters"
	provisioningActions "github.com/rancher/tests/actions/provisioning"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	cleanup "github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	nested "github.com/rancher/tfp-automation/tests/extensions/nestedModules"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/rancher2/resources"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestTfpProvisionDataDirectory(t *testing.T) {
	t.Parallel()

	p := resources.Setup(t)

	dataDirectories := config.TerraformConfig{
		DataDirectories: &config.DataDirectories{
			SystemAgentPath:  "/custom/agent",
			ProvisioningPath: "/custom/provisioning",
			K8sDistroPath:    "/custom/rke2",
		},
	}

	tests := []struct {
		name            string
		module          string
		nodeRoles       []config.Nodepool
		dataDirectories config.TerraformConfig
	}{
		{"RKE2_Data_Directory", p.RKE2Module, p.NodeRoles, dataDirectories},
		{"K3S_Data_Directory", p.K3SModule, p.NodeRoles, dataDirectories},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(p.CattleConfig)
			rancher.AdminToken = p.StandardToken
			terratest.Nodepools = tt.nodeRoles
			terraform.DataDirectories = tt.dataDirectories.DataDirectories
			terraform.Module = tt.module

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(p.TerraformConfig, p.TerratestConfig, p.TerraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer func() {
				resources.RemoveDirectory(t, p.RancherConfig, nestedRancherModuleDir)
			}()

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()

			terratest, err = provisioning.GetK8sVersion(p.Client, terraform, terratest)
			require.NoError(t, err)

			terraform = provisioning.UniquifyTerraform(terraform)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, p.TerratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(t, perTestTerraformOptions, keyPath)

			logrus.Infof("Provisioning cluster (%s)", terraform.ResourcePrefix)
			clusters, _ := provisioning.Provision(t, p.Client, p.StandardUserClient, rancher, terraform, terratest, perTestTerraformOptions, newFile, rootBody, file, false, false, false, "", nestedRancherModuleDir)

			logrus.Infof("Verifying the cluster is ready (%s)", clusters[0].Name)
			err = provisioningActions.VerifyClusterReady(p.Client, clusters[0])
			require.NoError(t, err)

			logrus.Infof("Verifying service account token secret (%s)", clusters[0].Name)
			err = clusterActions.VerifyServiceAccountTokenSecret(p.Client, clusters[0].Name)
			require.NoError(t, err)

			logrus.Infof("Verifying cluster pods (%s)", clusters[0].Name)
			err = pods.VerifyClusterPods(p.Client, clusters[0])
			require.NoError(t, err)

			params := tfpQase.GetProvisioningSchemaParams(p.TerraformConfig, p.TerratestConfig)
			err = qase.UpdateSchemaParameters(tt.name, params)
			if err != nil {
				logrus.Warningf("Failed to upload schema parameters %s", err)
			}
		})
	}

	if p.TerratestConfig.LocalQaseReporting {
		results.ReportTest(p.TerratestConfig)
	}
}
