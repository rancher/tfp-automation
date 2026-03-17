package provisioning

import (
	"os"
	"os/exec"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/stretchr/testify/require"
)

// GoogleDriverImport is a function that will run terraform import for the Google driver.
func GoogleDriverImport(t *testing.T, terraformOptions *terraform.Options) {
	cmd := exec.Command("terraform", "-chdir="+terraformOptions.TerraformDir, "import", rancher2.NodeDriver+"."+rancher2.NodeDriver, "google")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		require.NoError(t, err)
	}
}
