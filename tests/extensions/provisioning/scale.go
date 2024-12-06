package provisioning

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	framework "github.com/rancher/tfp-automation/framework/set"
	"github.com/stretchr/testify/require"
)

// Scale is a function that will run terraform apply and scale the provisioned
// cluster, according to user's desired amount.
func Scale(t *testing.T, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	testUser, testPassword, clusterName, poolName string, terraformOptions *terraform.Options) {
	err := framework.ConfigTF(nil, rancherConfig, terraformConfig, terratestConfig, testUser, testPassword, clusterName, poolName, "", nil)
	require.NoError(t, err)

	terraform.Apply(t, terraformOptions)
}
