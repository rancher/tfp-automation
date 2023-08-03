package functions

import (
	"os"
	"testing"

	framework "github.com/rancher/rancher/tests/framework/pkg/config"
	set "github.com/rancher/tfp-automation/functions/set"
	"github.com/rancher/tfp-automation/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BuildModule(t *testing.T) (bool, error) {
	clusterConfig := new(config.TerratestConfig)
	framework.LoadConfig("terratest", clusterConfig)

	keyPath := set.SetKeyPath()

	result, err := set.SetConfigTF(t, clusterConfig.KubernetesVersion, clusterConfig.Nodepools)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	module, err := os.ReadFile(keyPath + "/main.tf")

	if err != nil {
		t.Logf("Failed to read/grab main.tf file contents. Error: %v", err)
		return false, err
	}

	t.Log(string(module))

	return true, nil
}
