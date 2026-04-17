package provisioning

import (
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	steveV1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/tests/actions/clusters"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/providers"
	framework "github.com/rancher/tfp-automation/framework/set"
	"github.com/stretchr/testify/require"
)

// Provision is a function that will run terraform init and apply Terraform resources to provision a cluster.
func Provision(t *testing.T, client, standardUserClient *rancher.Client, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, testUser, testPassword string, terraformOptions *terraform.Options,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File, isWindows, persistClusters,
	containsCustomModule bool, clusterIDs []string, customClusterName string, nestedRancherModuleDir string) ([]*steveV1.SteveAPIObject, string) {
	var err error
	var clusterNames []string

	isSupported := SupportedModules(terraformConfig)
	require.True(t, isSupported)

	clusterNames, customClusterName, err = framework.ConfigTF(standardUserClient, rancherConfig, terratestConfig, testUser, testPassword, "", terraformConfig, newFile, rootBody, file, isWindows, persistClusters, containsCustomModule, customClusterName, nestedRancherModuleDir)
	require.NoError(t, err)

	// If the provisioner is GKE, we need to run terraform import for the Google driver before applying the Terraform configuration.
	// This is needed as the Google driver is inactive by default and needs to be imported to be activated.
	if terraformConfig.Module == providers.GKE || strings.Contains(terraformConfig.Module, "google") {
		terraform.Init(t, terraformOptions)
		GoogleDriverImport(t, terraformOptions)
	}

	terraform.InitAndApply(t, terraformOptions)

	var clusterObjects []*steveV1.SteveAPIObject
	for _, clusterName := range clusterNames {
		createdCluster, err := clusters.GetClusterByName(client, clusterName)
		require.NoError(t, err)

		clusterObjects = append(clusterObjects, createdCluster)
	}

	return clusterObjects, customClusterName
}
