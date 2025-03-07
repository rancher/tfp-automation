package provisioning

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	clusterExtensions "github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/tfp-automation/config"
	framework "github.com/rancher/tfp-automation/framework/set"
	"github.com/stretchr/testify/require"
)

// Provision is a function that will run terraform init and apply Terraform resources to provision a cluster.
func Provision(t *testing.T, client *rancher.Client, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, testUser, testPassword string, terraformOptions *terraform.Options, configMap []map[string]any,
	isWindows bool) []string {
	var err error
	var clusterNames []string
	var clusterIDs []string

	isSupported := SupportedModules(terraformOptions, configMap)
	require.True(t, isSupported)

	clusterNames, err = framework.ConfigTF(client, testUser, testPassword, "", configMap, isWindows)
	require.NoError(t, err)

	terraform.InitAndApply(t, terraformOptions)

	for _, clusterName := range clusterNames {
		clusterID, err := clusterExtensions.GetClusterIDByName(client, clusterName)
		require.NoError(t, err)

		clusterIDs = append(clusterIDs, clusterID)
	}

	return clusterIDs
}
