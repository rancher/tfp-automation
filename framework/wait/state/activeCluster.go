package state

import (
	"context"
	"time"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/defaults/clusterstate"
	"github.com/sirupsen/logrus"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// IsActiveCluster is a function that will wait for the cluster to be in an active state.
func IsActiveCluster(client *rancher.Client, clusterID string) error {
	err := kwait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 60*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		cluster, err := client.Management.Cluster.ByID(clusterID)
		if err != nil {
			return false, err
		}

		if cluster.State == clusterstate.ActiveState {
			logrus.Infof("Cluster %v is now active!", cluster.Name)
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return err
	}

	return nil
}
