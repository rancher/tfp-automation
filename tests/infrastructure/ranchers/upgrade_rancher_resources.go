package ranchers

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	extClusters "github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	provisioningActions "github.com/rancher/tests/actions/provisioning"
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// SetupResources is a function that creates a standard user and initializes the main.tf file for provisioning downstream clusters.
// This function is specifically designed for the upgrading Rancher tests to provision clusters before and after
// the Rancher upgrade.
func SetupResources(t *testing.T, client *rancher.Client, rancherConfig *rancher.Config, terratestConfig *config.TerratestConfig,
	terraformOptions *terraform.Options) (*rancher.Client, string, string, string) {
	var err error
	var testUser, testPassword string

	standardUserClient, testUser, testPassword, err := standarduser.CreateStandardUser(client)
	require.NoError(t, err)

	standardUserToken, err := CreateStandardUserToken(t, terraformOptions, rancherConfig, testUser, testPassword)
	require.NoError(t, err)

	standardToken := standardUserToken.Token

	return standardUserClient, standardToken, testUser, testPassword
}

// CleanupDownstreamClusters is a function that cleans up any downstream clusters created during the Rancher upgrade tests.
func CleanupDownstreamClusters(t *testing.T, client *rancher.Client, terraformConfig *config.TerraformConfig) {
	clusters, err := client.Steve.SteveType(stevetypes.Provisioning).ListAll(nil)
	require.NoError(t, err)

	for _, cluster := range clusters.Data {
		if cluster.Name == "local" {
			continue
		}

		err = provisioningActions.VerifyClusterReady(client, &cluster)
		require.NoError(t, err)

		dsCluster, err := client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + cluster.Name)
		require.NoError(t, err)

		logrus.Infof("Cleaning up cluster: %v", dsCluster.ID)
		err = extClusters.DeleteK3SRKE2Cluster(client, dsCluster.ID)
		require.NoError(t, err)

		provisioningActions.VerifyDeleteRKE2K3SCluster(t, client, dsCluster.ID)
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
