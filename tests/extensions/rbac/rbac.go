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
func RBAC(t *testing.T, client *rancher.Client, clusterName, poolName string, terraformOptions *terraform.Options,
	clusterConfig *config.TerratestConfig, rbacRole config.Role) {
	err := framework.ConfigTF(client, clusterConfig, clusterName, poolName, rbacRole)
	require.NoError(t, err)

	terraform.Apply(t, terraformOptions)
}
