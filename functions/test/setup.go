package functions

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	framework "github.com/rancher/rancher/tests/framework/pkg/config"
	set "github.com/josh-diamond/rancher-terratest-framework/functions/set"
	"github.com/josh-diamond/rancher-terratest-framework/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Setup(t *testing.T) (*terraform.Options, bool, error) {
	clusterConfig := new(config.TerratestConfig)
	framework.LoadConfig("terratest", clusterConfig)

	keyPath := set.SetKeyPath()

	result, err := set.SetConfigTF(t, clusterConfig.KubernetesVersion, clusterConfig.Nodepools)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{

		TerraformDir: keyPath,
		NoColor:      true,
	})

	return terraformOptions, true, nil
}
