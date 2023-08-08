package functions

import (
	"testing"
	"time"

	"github.com/rancher/norman/types"
	"github.com/rancher/rancher/tests/framework/clients/rancher"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/wait"
)

func ActiveNodes(t *testing.T, client *rancher.Client, clusterID string) (done bool, err error) {
	nodes, listErr := client.Management.Node.List(&types.ListOpts{
		Filters: map[string]interface{}{
			"clusterId": clusterID,
		},
	})
	require.NoError(t, err)

	if listErr != nil {
		t.Logf("Failed to locate cluster and grab node list.")
		return false, err
	}

	for _, node := range nodes.Data {

		if node.State != "active" {
			waiting := false
			waitErr := wait.Poll(500*time.Millisecond, 30*time.Minute, func() (done bool, err error) {
				n, clientErr := client.Management.Node.ByID(node.ID)
				require.NoError(t, err)

				if clientErr != nil {
					t.Logf("Failed to locate cluster and grab client specs. Error: %v", err)
					return false, err
				}

				if n.State == "active" {
					t.Logf("Node %v is now active.", n.Name)
					return true, nil
				}

				if !waiting {
					t.Logf("Waiting for cluster nodes to be in an active state...")
					waiting = true
				}

				return false, nil
			})
			require.NoError(t, waitErr)

			if waitErr != nil {
				t.Logf("Failed to instantiate active nodes wait poll.")
				return false, waitErr
			}
		}
	}

	return true, nil
}
