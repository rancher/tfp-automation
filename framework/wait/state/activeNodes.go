package state

import (
	"context"
	"time"

	"github.com/rancher/norman/types"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/tfp-automation/defaults/clusterstate"
	"github.com/sirupsen/logrus"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// AreNodesActive is a function that will wait for all nodes in the cluster to be in an active state.
func AreNodesActive(client *rancher.Client, clusterID string) error {
	err := kwait.PollUntilContextTimeout(context.TODO(), 10*time.Second, defaults.ThirtyMinuteTimeout, true, func(ctx context.Context) (done bool, err error) {
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

			if node.State == clusterstate.ErrorState {
				logrus.Warnf("Node %s is in error state", node.Name)
				return false, nil
			}

			if node.State != clusterstate.ActiveState {
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
