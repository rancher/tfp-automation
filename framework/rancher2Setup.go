package framework

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	setFramework "github.com/rancher/tfp-automation/framework/set"
	resources "github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/stretchr/testify/require"
)

// Rancher2Setup is a function that will set the Terraform configuration and return the Terraform options.
func Rancher2Setup(t *testing.T, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig) *terraform.Options {
	keyPath := resources.SetKeyPath()

	err := setFramework.ConfigTF(nil, rancherConfig, terraformConfig, terratestConfig, "", "", "", "", "", nil)
	require.NoError(t, err)

	var terratestLogger logger.Logger
	if terratestConfig.TFLogging {
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
