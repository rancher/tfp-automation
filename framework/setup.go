package framework

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/tfp-automation/config"
)

// Setup is a function that will set the Terraform configuration and return the Terraform options.
func Setup(t *testing.T, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, keyPath string) *terraform.Options {
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
