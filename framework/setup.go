package framework

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	framework "github.com/rancher/shepherd/pkg/config"
	"github.com/josh-diamond/tfp-automation/config"
	set "github.com/josh-diamond/tfp-automation/framework/set"
)

const (
	terratest                = "terratest"
	terraformFrameworkConfig = "terraform"
)

// Setup is a function that will set the Terraform configuration and return the Terraform options.
func Setup(t *testing.T) *terraform.Options {
	clusterConfig := new(config.TerratestConfig)
	framework.LoadConfig(terratest, clusterConfig)

	keyPath := set.SetKeyPath()

	set.SetConfigTF(clusterConfig, "")

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: keyPath,
		NoColor:      true,
	})

	return terraformOptions
}
