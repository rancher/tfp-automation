package provisioning

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/josh-diamond/tfp-automation/config"
	set "github.com/josh-diamond/tfp-automation/framework/set"
	"github.com/stretchr/testify/require"
)

// KubernetesUpgrade is a function that will run terraform apply and uprade the Kubernetes version of the provisioned cluster.
func KubernetesUpgrade(t *testing.T, clusterName string, terraformOptions *terraform.Options, clusterConfig *config.TerratestConfig) {
	clusterConfig.KubernetesVersion = clusterConfig.UpgradedKubernetesVersion

	err := set.SetConfigTF(clusterConfig, clusterName)
	require.NoError(t, err)

	terraform.Apply(t, terraformOptions)
}
