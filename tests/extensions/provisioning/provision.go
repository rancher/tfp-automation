package provisioning

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	clusterExtensions "github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/tfp-automation/config"
	framework "github.com/rancher/tfp-automation/framework/set"
	"github.com/stretchr/testify/require"
)

// Provision is a function that will run terraform init and apply Terraform resources to provision a cluster.
func Provision(t *testing.T, client, standardUserClient *rancher.Client, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, testUser, testPassword string, terraformOptions *terraform.Options,
	configMap []map[string]any, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File, isWindows, persistClusters,
	containsCustomModule bool, clusterIDs, customClusterNames []string) ([]string, []string) {
	var err error
	var clusterNames []string

	isSupported := SupportedModules(terraformOptions, configMap)
	require.True(t, isSupported)

	clusterNames, customClusterNames, err = framework.ConfigTF(standardUserClient, rancherConfig, terratestConfig, testUser, testPassword, "", configMap, newFile, rootBody, file, isWindows, persistClusters, containsCustomModule, customClusterNames)
	require.NoError(t, err)

	terraform.InitAndApply(t, terraformOptions)

	for _, clusterName := range clusterNames {
		clusterID, err := clusterExtensions.GetClusterIDByName(client, clusterName)
		require.NoError(t, err)

		clusterIDs = append(clusterIDs, clusterID)
	}

	return clusterIDs, customClusterNames
}
