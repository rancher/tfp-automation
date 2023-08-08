package functions

import (
	"testing"
	"time"

	"github.com/rancher/rancher/tests/framework/clients/rancher"
	framework "github.com/rancher/rancher/tests/framework/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/wait"
)

func ScaleUp(t *testing.T, client *rancher.Client, clusterID string) (done bool, err error) {
	clusterConfig := new(config.TerratestConfig)
	framework.LoadConfig("terratest", clusterConfig)
	waiting := false
	waitErr := wait.Poll(100*time.Millisecond, 30*time.Minute, func() (done bool, err error) {
		cluster, clientErr := client.Management.Cluster.ByID(clusterID)
		require.NoError(t, err)

		if clientErr != nil {
			t.Logf("Failed to locate cluster and grab client specs. Error: %v", err)
			return false, err
		}

		if cluster.NodeCount == clusterConfig.ScaledUpNodeCount {
			t.Logf("Successfully scaled up cluster to %v nodes", clusterConfig.ScaledUpNodeCount)
			return true, nil
		}

		if !waiting {
			t.Logf("Waiting for cluster to scale up...")
			waiting = true
		}
		return false, nil
	})
	require.NoError(t, waitErr)

	if waitErr != nil {
		t.Logf("Failed to instantiate scale up wait poll.")
		return false, waitErr
	}

	return true, nil
}
