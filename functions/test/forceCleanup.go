package functions

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	cleanup "github.com/josh-diamond/rancher-terratest-framework/functions/cleanup"
	set "github.com/josh-diamond/rancher-terratest-framework/functions/set"
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
