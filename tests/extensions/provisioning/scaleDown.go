package provisioning

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/josh-diamond/tfp-automation/config"
	set "github.com/josh-diamond/tfp-automation/framework/set"
	"github.com/stretchr/testify/require"
)

// ScaleDown is a function that will run terraform apply and scale down the provisioned cluster.
func ScaleDown(t *testing.T, clusterName string, terraformOptions *terraform.Options, clusterConfig *config.TerratestConfig) {
	clusterConfig.Nodepools = clusterConfig.ScaledDownNodepools

	err := set.SetConfigTF(clusterConfig, clusterName)
	require.NoError(t, err)

	terraform.Apply(t, terraformOptions)
}
