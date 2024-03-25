package framework

import (
	"time"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	activeState   = "active"
	errorState    = "error"
	True          = "True"
	waitingState  = "waiting"
	updatingState = "updating"
)

// IsActiveCluster is a function that will wait for the cluster to be in an active state.
func IsActiveCluster(client *rancher.Client, clusterID string) error {
	isWaiting := true
	err := wait.Poll(10*time.Second, defaults.ThirtyMinuteTimeout, func() (done bool, err error) {
		cluster, err := client.Management.Cluster.ByID(clusterID)
		if err != nil {
			return false, err
		}

		if cluster.State == activeState {
			logrus.Infof("Cluster %v is now active!", cluster.Name)
			return true, nil
		}

		if isWaiting {
			logrus.Infof("Waiting for cluster %v to be in an active state...", cluster.Name)
		}

		return false, nil
	})
	if err != nil {
		return err
	}

	return nil
}
