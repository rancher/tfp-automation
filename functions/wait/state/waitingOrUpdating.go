package functions

import (
	"testing"
	"time"

	"github.com/rancher/rancher/tests/framework/clients/rancher"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/wait"
)

func WaitingOrUpdating(t *testing.T, client *rancher.Client, clusterID string) (done bool, err error) {
	waiting := false
	waitErr := wait.Poll(100*time.Millisecond, 30*time.Minute, func() (done bool, err error) {
		cluster, clientErr := client.Management.Cluster.ByID(clusterID)
		require.NoError(t, err)

		if clientErr != nil {
			t.Logf("Failed to locate cluster and grab client specs. Error: %v", err)
			return false, err
		}

		if cluster.State == "waiting" {
			t.Logf("Cluster is now in a waiting state.")
			return true, nil
		}

		if cluster.State == "updating" {
			t.Logf("Cluster is now in an updating state.")
			return true, nil
		}

		if !waiting {
			t.Logf("Waiting for cluster nodes to be in waiting or updating state...")
			waiting = true
		}

		return false, nil
	})
	require.NoError(t, waitErr)

	if waitErr != nil {
		t.Logf("Failed to instantiate waiting or updating wait poll.")
		return false, waitErr
	}

	return true, nil
}
