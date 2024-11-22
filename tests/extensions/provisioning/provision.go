package provisioning

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	framework "github.com/rancher/tfp-automation/framework/set"
	"github.com/stretchr/testify/require"
)

// Provision is a function that will run terraform init and apply Terraform resources to provision a cluster.
func Provision(t *testing.T, client *rancher.Client, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	clusterConfig *config.TerratestConfig, testUser, testPassword, clusterName, poolName string, terraformOptions *terraform.Options) {
	isSupported := SupportedModules(terraformConfig, terraformOptions)
	require.True(t, isSupported)

	err := framework.ConfigTF(client, rancherConfig, terraformConfig, clusterConfig, testUser, testPassword, clusterName, poolName, "")
	require.NoError(t, err)

	terraform.InitAndApply(t, terraformOptions)
}
