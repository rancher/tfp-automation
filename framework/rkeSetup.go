package framework

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/tfp-automation/config"
	resources "github.com/rancher/tfp-automation/framework/set/resources/rke"
)

// RKESetup is a function that will set the Terraform configuration and return the Terraform options.
func RKESetup(t *testing.T, terraformConfig *config.TerraformConfig) *terraform.Options {
	keyPath := resources.KeyPath()

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: keyPath,
		NoColor:      true,
	})

	return terraformOptions
}
