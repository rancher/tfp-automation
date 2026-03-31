package custom

import (
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	v1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	nested "github.com/rancher/tfp-automation/tests/extensions/nestedModules"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
)

// CreateCustomCluster is a function that will create a custom cluster using the provided terraform configuration and return the cluster object and the terraform options used to create the cluster.
func CreateCustomCluster(t *testing.T, client *rancher.Client, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, clusterType, dataDir string) (string, *terraform.Options, string, *v1.SteveAPIObject) {
	terraformConfig = provisioning.UniquifyTerraform(terraformConfig)
	terraformConfig.Module = fmt.Sprintf("%s_%s_custom", terraformConfig.DownstreamClusterProvider, clusterType)

	_, keyPath := rancher2.SetKeyPath(dataDir, terratestConfig.PathToRepo, terraformConfig.Provider)
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(terraformConfig, terratestConfig, terraformOptions, t.Name(), dataDir)
	require.NoError(t, err)

	newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
	defer file.Close()

	adminClient, err := rancher.NewClient("", client.Session)
	require.NoError(t, err)

	terratestConfig, err = provisioning.GetK8sVersion(adminClient, terraformConfig, terratestConfig)
	require.NoError(t, err)

	clusters, _ := provisioning.Provision(t, adminClient, client, rancherConfig, terraformConfig, terratestConfig, "", "", perTestTerraformOptions,
		newFile, rootBody, file, false, false, true, "", nestedRancherModuleDir)
	require.NotEmpty(t, clusters)

	return nestedRancherModuleDir, perTestTerraformOptions, keyPath, clusters[0]
}
