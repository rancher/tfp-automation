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

// KubernetesUpgrade is a function that will run terraform apply and uprade the
// Kubernetes version of the provisioned cluster.
func KubernetesUpgrade(t *testing.T, client *rancher.Client, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, testUser, testPassword string, terraformOptions *terraform.Options, configMap []map[string]any,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File, isWindows bool) ([]string, []string) {
	var err error
	var clusterNames []string
	var clusterIDs []string

	DefaultUpgradedK8sVersion(t, client, terratestConfig, terraformConfig, configMap)

	clusterNames, customClusterNames, err := framework.ConfigTF(client, testUser, testPassword, "", configMap, newFile, rootBody, file, isWindows, false, false, nil)
	require.NoError(t, err)

	terraform.Apply(t, terraformOptions)

	for _, clusterName := range clusterNames {
		clusterID, err := clusterExtensions.GetClusterIDByName(client, clusterName)
		require.NoError(t, err)

		clusterIDs = append(clusterIDs, clusterID)
	}

	return clusterIDs, customClusterNames
}
