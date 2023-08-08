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

func ScaleDown(t *testing.T, client *rancher.Client, clusterID string) (done bool, err error) {
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

		if cluster.NodeCount == clusterConfig.ScaledDownNodeCount {
			t.Logf("Successfully scaled down cluster to %v nodes", clusterConfig.ScaledDownNodeCount)
			return true, nil
		}

		if !waiting {
			t.Logf("Waiting for cluster to scale down...")
			waiting = true
		}

		return false, nil
	})
	require.NoError(t, waitErr)

	if waitErr != nil {
		t.Logf("Failed to instantiate scale down wait poll.")
		return false, waitErr
	}

	return true, nil
}
