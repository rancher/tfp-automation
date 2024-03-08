package provisioning

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/tfp-automation/config"
	set "github.com/rancher/tfp-automation/framework/set/provisioning"
	"github.com/stretchr/testify/require"
)

// Scale is a function that will run terraform apply and scale the provisioned
// cluster, according to user's desired amount.
func Scale(t *testing.T, clusterName string, terraformOptions *terraform.Options, clusterConfig *config.TerratestConfig) {
	err := set.SetConfigTF(clusterConfig, clusterName)
	require.NoError(t, err)

	terraform.Apply(t, terraformOptions)
}
