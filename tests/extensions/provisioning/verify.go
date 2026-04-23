package provisioning

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tests/actions/registries"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// VerifyRegistry validates that the expected registry is set.
func VerifyRegistry(t *testing.T, client *rancher.Client, clusterID string, terraformConfig *config.TerraformConfig) {
	if terraformConfig.PrivateRegistries != nil {
		_, err := registries.CheckAllClusterPodsForRegistryPrefix(client, clusterID, terraformConfig.PrivateRegistries.URL)
		require.NoError(t, err)
	}
}

// VerifyRancherVersion validates that the expected rancher version matches the version of the rancher server.
func VerifyRancherVersion(t *testing.T, hostURL, expectedVersion, keyPath string, terraformOptions *terraform.Options) {
	resp, err := RequestRancherVersion(hostURL)
	require.NoError(t, err)

	logrus.Infof("Rancher version: %s | Rancher commit: %s", resp.RancherVersion, resp.GitCommit)

	if resp.RancherVersion != expectedVersion {
		logrus.Infof("Expected version: %s | Actual version: %s", expectedVersion, resp.RancherVersion)
		cleanup.Cleanup(t, terraformOptions, keyPath)
	}
}
