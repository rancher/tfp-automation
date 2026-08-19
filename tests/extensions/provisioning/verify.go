package provisioning

import (
	"regexp"
	"strings"
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
func VerifyRancherVersion(t *testing.T, hostURL, expectedVersion, keyPath string, terraformConfig *config.TerraformConfig,
	terraformOptions *terraform.Options) {
	resp, err := RequestRancherVersion(hostURL)
	require.NoError(t, err)

	logrus.Infof("Rancher version: %s | Rancher commit: %s", resp.RancherVersion, resp.GitCommit)

	if strings.Contains(terraformConfig.Standalone.Repo, "prime-release") || strings.Contains(terraformConfig.Standalone.UpgradedRancherRepo, "prime-release") {
		expectedVersionPrefix := strings.TrimSuffix(expectedVersion, "-head")
		expectedVersionPrefix = strings.TrimSuffix(expectedVersionPrefix, ".x")
		versionPattern := "^" + regexp.QuoteMeta(expectedVersionPrefix) + `\.[0-9]+-head$`

		if matched, regexErr := regexp.MatchString(versionPattern, resp.RancherVersion); regexErr == nil && matched {
			expectedVersion = resp.RancherVersion
		}
	}

	if resp.RancherVersion != expectedVersion {
		logrus.Infof("Expected version: %s | Actual version: %s", expectedVersion, resp.RancherVersion)
		cleanup.Cleanup(t, terraformOptions, keyPath)
	}
}
