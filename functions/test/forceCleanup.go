package functions

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	cleanup "github.com/rancher/tfp-automation/functions/cleanup"
	set "github.com/rancher/tfp-automation/functions/set"
)

func ForceCleanup(t *testing.T) (bool, error) {

	keyPath := set.SetKeyPath()

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: keyPath,
		NoColor:      true,
	})

	terraform.Destroy(t, terraformOptions)
	cleanup.CleanupConfigTF(t)

	return true, nil
}
