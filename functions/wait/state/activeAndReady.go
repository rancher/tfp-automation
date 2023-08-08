package functions

import (
	"testing"
	"time"

	"github.com/rancher/rancher/tests/framework/clients/rancher"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/wait"
)

func ActiveAndReady(t *testing.T, client *rancher.Client, clusterID string) (done bool, err error) {
	waiting := false
	waitErr := wait.Poll(100*time.Millisecond, 30*time.Minute, func() (done bool, err error) {
		cluster, clientErr := client.Management.Cluster.ByID(clusterID)
		require.NoError(t, err)

		if clientErr != nil {
			t.Logf("Failed to locate cluster and grab client specs. Error: %v", err)
			return false, err
		}

		if cluster.State == "active" && cluster.Conditions[0].Status == "True" {
			t.Logf("Cluster is now active and ready.")
			return true, nil
		}
		
		if !waiting {
			t.Logf("Waiting for cluster to be in an active and ready state...")
			waiting = true
		}

		return false, nil
	})
	require.NoError(t, waitErr)

	if waitErr != nil {
		t.Logf("Failed to instantiate active and ready wait poll.")
		return false, waitErr
	}

	return true, nil
}
