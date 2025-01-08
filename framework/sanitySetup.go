package framework

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/tfp-automation/config"
	resources "github.com/rancher/tfp-automation/framework/set/resources/sanity"
)

// SanitySetup is a function that will set the Terraform configuration and return the Terraform options.
func SanitySetup(t *testing.T, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig) (*terraform.Options, string) {
	keyPath := resources.KeyPath()

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

	return terraformOptions, keyPath
}
