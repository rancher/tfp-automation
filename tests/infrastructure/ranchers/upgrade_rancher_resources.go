package ranchers

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
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
func CleanupPreUpgradeClusters(t *testing.T, client *rancher.Client, terraformOptions *terraform.Options, terratestConfig *config.TerratestConfig,
	cattleConfig map[string]any, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) {
	_, err := operations.ReplaceValue([]string{"rancher", "adminToken"}, client.RancherConfig.AdminToken, cattleConfig)
	require.NoError(t, err)

	_, _, terratestConfig, _ = config.LoadTFPConfigs(cattleConfig)

	newFile, rootBody, file = rancher2.InitializeMainTF(terratestConfig)
	defer file.Close()

	_, err = file.Write(newFile.Bytes())
	require.NoError(t, err)

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	cleanup.Cleanup(t, terraformOptions, keyPath)
}
