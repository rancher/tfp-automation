package provisioning

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	cleanup "github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources"
)

// ForceCleanup is a function that will forcibly run terraform destroy and cleanup Terraform resources.
func ForceCleanup(t *testing.T) {
	keyPath := resources.SetKeyPath()

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: keyPath,
		NoColor:      true,
	})

	terraform.Destroy(t, terraformOptions)
	cleanup.CleanupConfigTF()
}
