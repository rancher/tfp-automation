package provisioning

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
)

// ForceCleanup is a function that will forcibly run terraform destroy and cleanup Terraform resources.
func ForceCleanup(t *testing.T) error {
	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, "")

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: keyPath,
		NoColor:      true,
	})

	terraform.Destroy(t, terraformOptions)
	cleanup.TFFilesCleanup(keyPath)

	return nil
}
