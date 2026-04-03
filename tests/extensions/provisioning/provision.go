package provisioning

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	steveV1 "github.com/rancher/shepherd/clients/rancher/v1"
	extClusters "github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/providers"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
	framework "github.com/rancher/tfp-automation/framework/set"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// Provision is a function that will run terraform init and apply Terraform resources to provision a cluster.
func Provision(t *testing.T, client, standardUserClient *rancher.Client, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, testUser, testPassword string, terraformOptions *terraform.Options,
	configMap []map[string]any, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File, isWindows, persistClusters,
	containsCustomModule bool, clusterIDs, customClusterNames []string, nestedRancherModuleDir string) ([]*steveV1.SteveAPIObject, []string) {
	var err error
	var clusterNames []string

	isSupported := SupportedModules(terraformOptions, configMap)
	require.True(t, isSupported)

	clusterNames, customClusterNames, err = framework.ConfigTF(standardUserClient, rancherConfig, terratestConfig, testUser, testPassword, "", configMap, newFile, rootBody, file, isWindows, persistClusters, containsCustomModule, customClusterNames, nestedRancherModuleDir)
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
		var createdCluster *steveV1.SteveAPIObject
		err = kwait.PollUntilContextTimeout(context.TODO(), 10*time.Second, defaults.FiveMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
			adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
			if err != nil {
				logrus.Warningf("Unable to get admin cluster client (%s) retrying", clusterName)
				return false, nil
			}

			if IsImportedModule(terraformConfig.Module) || IsHostedModule(terraformConfig.Module) {
				clusterName, err = extClusters.GetClusterIDByName(adminClient, terraformConfig.ResourcePrefix)
				if err != nil {
					return false, nil
				}
			}

			createdCluster, err = adminClient.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + clusterName)
			if err != nil {
				logrus.Warningf("Unable to get cluster (%s) retrying", clusterName)
				return false, nil
			}

			return true, nil
		})
		require.NoError(t, err)

		clusterObjects = append(clusterObjects, createdCluster)
	}

	return clusterObjects, customClusterNames
}
