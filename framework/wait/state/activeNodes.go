package framework

import (
	"time"

	"github.com/rancher/norman/types"
	"github.com/rancher/rancher/tests/framework/clients/rancher"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

// AreNodesActive is a function that will wait for all nodes in the cluster to be in an active state.
func AreNodesActive(client *rancher.Client, clusterID string) error {
	err := wait.Poll(10*time.Second, 10*time.Minute, func() (bool, error) {
		nodes, err := client.Management.Node.ListAll(&types.ListOpts{
			Filters: map[string]interface{}{
				"clusterId": clusterID,
			},
		})
		if err != nil {
			return false, nil
		}

		for _, node := range nodes.Data {
			node, err := client.Management.Node.ByID(node.ID)
			if err != nil {
				return false, nil
			}

			if node.State == errorState {
				logrus.Warnf("Node %s is in error state", node.Name)
				return false, nil
			}

			if node.State != activeState {
				return false, nil
			}
		}

		logrus.Infof("All nodes in the cluster are in an active state!")
		return true, nil
	})
	if err != nil {
		return err
	}

	return nil
}
