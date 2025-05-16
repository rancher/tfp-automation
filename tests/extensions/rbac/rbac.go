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

// RBAC is a function that will run terraform apply to create users.
func RBAC(t *testing.T, client *rancher.Client, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, testUser, testPassword string, terraformOptions *terraform.Options,
	configMap []map[string]any, rbacRole config.Role, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) {
	_, _, err := framework.ConfigTF(client, rancherConfig, testUser, testPassword, rbacRole, configMap, newFile, rootBody, file, false, false, false, nil)
	require.NoError(t, err)

	terraform.Apply(t, terraformOptions)
}
