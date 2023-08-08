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

func KubernetesUpgrade(t *testing.T, client *rancher.Client, clusterID string, module string) (done bool, err error) {
	clusterConfig := new(config.TerratestConfig)
	framework.LoadConfig("terratest", clusterConfig)

	if module == "aks" || module == "gke" {
		expectedUpgradedKubernetesVersion := `v` + clusterConfig.UpgradedKubernetesVersion
		waiting := false
		waitErr := wait.Poll(100*time.Millisecond, 30*time.Minute, func() (done bool, err error) {
			cluster, clientErr := client.Management.Cluster.ByID(clusterID)
			require.NoError(t, err)

			if clientErr != nil {
				t.Logf("Failed to locate cluster and grab client specs. Error: %v", err)
				return false, err
			}

			if cluster.Version.GitVersion == expectedUpgradedKubernetesVersion {
				t.Logf("Successfully updated kubernetes version to %v", clusterConfig.UpgradedKubernetesVersion)
				return true, nil
			}

			if !waiting {
				t.Logf("Waiting for cluster to upgrade kubernetes version...")
				waiting = true
			}

			return false, nil
		})
		require.NoError(t, waitErr)

		if waitErr != nil {
			t.Logf("Failed to instantiate kubernetes upgrade wait poll.")
			return false, waitErr
		}
	}

	if module == "ec2_rke1" || module == "linode_rke1" {
		expectedUpgradedKubernetesVersion := clusterConfig.UpgradedKubernetesVersion[:len(clusterConfig.UpgradedKubernetesVersion)-11]
		waiting := false
		waitErr := wait.Poll(100*time.Millisecond, 30*time.Minute, func() (done bool, err error) {
			cluster, clientErr := client.Management.Cluster.ByID(clusterID)
			require.NoError(t, err)

			if clientErr != nil {
				t.Logf("Failed to locate cluster and grab client specs. Error: %v", err)
				return false, err
			}

			if cluster.Version.GitVersion == expectedUpgradedKubernetesVersion {
				return true, nil
			}

			if !waiting {
				t.Logf("Waiting for cluster to upgrade kubernetes version...")
				waiting = true
			}

			return false, nil
		})
		require.NoError(t, waitErr)

		if waitErr != nil {
			t.Logf("Failed to instantiate kubernetes upgrade wait poll.")
			return false, waitErr
		}
	}

	if module == "ec2_k3s" || module == "ec2_rke2" || module == "linode_k3s" || module == "linode_rke2" {
		waiting := false
		waitErr := wait.Poll(100*time.Millisecond, 30*time.Minute, func() (done bool, err error) {
			cluster, clientErr := client.Management.Cluster.ByID(clusterID)
			require.NoError(t, err)

			if clientErr != nil {
				t.Logf("Failed to locate cluster and grab client specs. Error: %v", err)
				return false, err
			}

			if cluster.Version.GitVersion == clusterConfig.UpgradedKubernetesVersion {
				return true, nil
			}

			if !waiting {
				t.Logf("Waiting for cluster to upgrade kubernetes version...")
				waiting = true
			}
			
			return false, nil
		})
		require.NoError(t, waitErr)

		if waitErr != nil {
			t.Logf("Failed to instantiate kubernetes upgrade wait poll.")
			return false, waitErr
		}
	}

	return true, nil
}
