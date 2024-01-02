package provisioning

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/tfp-automation/config"
	set "github.com/rancher/tfp-automation/framework/set"
	"github.com/stretchr/testify/require"
)

// ScaleUp is a function that will run terraform apply and scale up the provisioned cluster.
func ScaleUp(t *testing.T, clusterName string, terraformOptions *terraform.Options, clusterConfig *config.TerratestConfig) {
	clusterConfig.Nodepools = clusterConfig.ScaledUpNodepools

	err := set.SetConfigTF(clusterConfig, clusterName)
	require.NoError(t, err)

	terraform.Apply(t, terraformOptions)
}
