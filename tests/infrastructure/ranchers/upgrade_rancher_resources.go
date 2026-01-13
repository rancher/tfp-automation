package ranchers

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	extClusters "github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	"github.com/rancher/tests/actions/provisioning"
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// SetupResources is a function that creates a standard user and initializes the main.tf file for provisioning downstream clusters.
// This function is specifically designed for the upgrading Rancher tests to provision clusters before and after
// the Rancher upgrade.
func SetupResources(t *testing.T, client *rancher.Client, rancherConfig *rancher.Config, terratestConfig *config.TerratestConfig,
	terraformOptions *terraform.Options) (*rancher.Client, *hclwrite.File, *hclwrite.Body, *os.File, string, string, string) {
	var err error
	var testUser, testPassword string

	newFile, rootBody, file := rancher2.InitializeMainTF(terratestConfig)
	defer file.Close()

	standardUserClient, testUser, testPassword, err := standarduser.CreateStandardUser(client)
	require.NoError(t, err)

	standardUserToken, err := CreateStandardUserToken(t, terraformOptions, rancherConfig, testUser, testPassword)
	require.NoError(t, err)

	standardToken := standardUserToken.Token

	return standardUserClient, newFile, rootBody, file, standardToken, testUser, testPassword
}

// CleanupPreUpgradeClusters is a function that cleans up any pre-upgrade downstream clusters created during the Rancher upgrade tests.
func CleanupPreUpgradeClusters(t *testing.T, client *rancher.Client, clusterIDs []string, terraformConfig *config.TerraformConfig) {
	for _, clusterID := range clusterIDs {
		clusterResp, err := client.Management.Cluster.ByID(clusterID)
		require.NoError(t, err)

		cluster, err := client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + clusterResp.Name)
		require.NoError(t, err)

		logrus.Infof("Cleaning up pre-upgrade cluster: %v", cluster.ID)
		err = extClusters.DeleteK3SRKE2Cluster(client, cluster.ID)
		require.NoError(t, err)

		provisioning.VerifyDeleteRKE2K3SCluster(t, client, cluster.ID)
	}
}

// UniqueStrings is a function that removes duplicate strings from a slice.
func UniqueStrings(clusterIDs []string) []string {
	duplicateName := make(map[string]struct{})
	var result []string

	for _, v := range clusterIDs {
		if _, ok := duplicateName[v]; !ok {
			duplicateName[v] = struct{}{}
			result = append(result, v)
		}
	}

	return result
}
