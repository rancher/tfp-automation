package rbac

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	framework "github.com/rancher/tfp-automation/framework/set"
	"github.com/stretchr/testify/require"
)

// AuthConfig is a function that will run terraform apply to setup authentication providers.
func AuthConfig(t *testing.T, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, terraformOptions *terraform.Options, testUser, testPassword string,
	configMap []map[string]any, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) {
	isSupported := SupportedAuthProviders(terraformConfig, terraformOptions)
	require.True(t, isSupported)

	err := framework.AuthConfig(rancherConfig, testUser, testPassword, configMap, newFile, rootBody, file)
	require.NoError(t, err)

	terraform.InitAndApply(t, terraformOptions)
}
