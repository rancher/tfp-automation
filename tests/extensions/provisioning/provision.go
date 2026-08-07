package provisioning

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/shell"
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
	terratestConfig *config.TerratestConfig, terraformOptions *terraform.Options,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File, isWindows, persistClusters,
	containsCustomModule bool, customClusterName string, nestedRancherModuleDir string) ([]*steveV1.SteveAPIObject, string) {
	var err error
	var clusterNames []string

	isSupported := SupportedModules(terraformConfig)
	require.True(t, isSupported)

	clusterNames, customClusterName, err = framework.ConfigTF(standardUserClient, rancherConfig, terratestConfig, "", terraformConfig, newFile, rootBody, file, isWindows, persistClusters, containsCustomModule, customClusterName, nestedRancherModuleDir)
	require.NoError(t, err)

	// If the provisioner is GKE, we need to run terraform import for the Google driver before applying the Terraform configuration.
	// This is needed as the Google driver is inactive by default and needs to be imported to be activated.
	if terraformConfig.Module == providers.GKE || strings.Contains(terraformConfig.Module, "google") {
		terraform.Init(t, terraformOptions)
		GoogleDriverImport(t, terraformOptions)
	}

	output, err := terraform.InitAndApplyE(t, terraformOptions)
	if err != nil {
		if output == "" {
			var fatalErr retry.FatalError
			var cmdErr *shell.ErrWithCmdOutput
			if errors.As(err, &fatalErr) {
				if errors.As(fatalErr.Underlying, &cmdErr) {
					output = cmdErr.Output.Combined()
				}
			}
		}
		require.NoError(t, err, "terraform apply failed. Output:\n%s", filterTerraformOutput(output))
	}

	var clusterObjects []*steveV1.SteveAPIObject
	for _, clusterName := range clusterNames {
		createdCluster, err := clusters.GetClusterByName(client, clusterName)
		require.NoError(t, err)

		clusterObjects = append(clusterObjects, createdCluster)
	}

	return clusterObjects, customClusterName
}

// filterTerraformOutput retains only error-relevant lines from terraform output,
func filterTerraformOutput(output string) string {
	var filtered []string
	for _, line := range strings.Split(output, "\n") {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "error") || strings.Contains(lower, "failed") {
			filtered = append(filtered, line)
		}
	}
	if len(filtered) == 0 {
		return output
	}
	return strings.Join(filtered, "\n")
}
