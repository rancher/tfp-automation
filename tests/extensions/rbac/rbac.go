package rbac

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	framework "github.com/rancher/tfp-automation/framework/set"
	"github.com/stretchr/testify/require"
)

// RBAC is a function that will run terraform apply to create users.
func RBAC(t *testing.T, client *rancher.Client, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, testUser, testPassword string, terraformOptions *terraform.Options,
	rbacRole config.Role) {
	_, err := framework.ConfigTF(client, testUser, testPassword, rbacRole, nil, false)
	require.NoError(t, err)

	terraform.Apply(t, terraformOptions)
}
