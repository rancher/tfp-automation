package framework

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	setFramework "github.com/rancher/tfp-automation/framework/set"
	"github.com/rancher/tfp-automation/framework/set/resources"
	"github.com/stretchr/testify/require"
)

// Setup is a function that will set the Terraform configuration and return the Terraform options.
func Setup(t *testing.T, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, clusterConfig *config.TerratestConfig) *terraform.Options {
	keyPath := resources.SetKeyPath()

	err := setFramework.ConfigTF(nil, rancherConfig, terraformConfig, clusterConfig, "", "", "", "", "", nil)
	require.NoError(t, err)

	var terratestLogger logger.Logger
	if clusterConfig.TFLogging {
		terratestLogger = *logger.Default
	} else {
		terratestLogger = *logger.Discard
	}

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: keyPath,
		NoColor:      true,
		Logger:       &terratestLogger,
	})

	return terraformOptions
}
