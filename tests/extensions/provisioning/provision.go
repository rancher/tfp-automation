package provisioning

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	set "github.com/rancher/tfp-automation/framework/set/provisioning"
	"github.com/stretchr/testify/require"
)

const (
	TFP = "tfp"
)

// Provision is a function that will run terraform init and apply Terraform resources to provision a cluster.
func Provision(t *testing.T, client *rancher.Client, clusterName, poolName string, clusterConfig *config.TerratestConfig, terraformOptions *terraform.Options) {
	err := set.SetConfigTF(clusterConfig, clusterName, poolName)
	require.NoError(t, err)

	isSupported := SupportedModules(terraformOptions)

	terraform.InitAndApply(t, terraformOptions)
	require.True(t, isSupported)
}
