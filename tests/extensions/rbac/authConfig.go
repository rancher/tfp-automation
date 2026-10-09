package rbac

import (
	"context"
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/providerversion"
	framework "github.com/rancher/tfp-automation/framework/set"
	"github.com/stretchr/testify/require"
)

// AuthConfig is a function that will run terraform apply to setup authentication providers.
func AuthConfig(t *testing.T, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, terraformOptions *terraform.Options,
	configMap []map[string]any, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) {
	isSupported := SupportedAuthProviders(terraformConfig, terraformOptions)
	require.True(t, isSupported)

	err := framework.AuthConfig(rancherConfig, configMap, newFile, rootBody, file)
	require.NoError(t, err)

	err = providerversion.ClearStaleLock(terraformOptions.TerraformDir)
	require.NoError(t, err)

	terraform.InitAndApplyContext(t, context.Background(), terraformOptions)
}
