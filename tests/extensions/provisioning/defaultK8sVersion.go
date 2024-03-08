package provisioning

import (
	"strings"
	"testing"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/clusters/kubernetesversions"
	"github.com/rancher/tfp-automation/config"
	"github.com/stretchr/testify/require"
)

const (
	RKE1    = "rke1"
	RKE2    = "rke2"
	K3S     = "k3s"
	RANCHER = "-rancher"
)

// DefaultK8sVersion is a function that will set the default Kubernetes version if the user has not specified one.
func DefaultK8sVersion(t *testing.T, client *rancher.Client, clusterConfig *config.TerratestConfig, terraformConfig *config.TerraformConfig) {
	if clusterConfig.KubernetesVersion == "" {
		if strings.Contains(terraformConfig.Module, RKE1) {
			defaultVersion, err := kubernetesversions.Default(client, clusters.RKE1ClusterType.String(), nil)
			require.NoError(t, err)

			clusterConfig.KubernetesVersion = defaultVersion[0]
		} else if strings.Contains(terraformConfig.Module, RKE2) {
			defaultVersion, err := kubernetesversions.Default(client, clusters.RKE2ClusterType.String(), nil)
			require.NoError(t, err)

			clusterConfig.KubernetesVersion = defaultVersion[0]
		} else if strings.Contains(terraformConfig.Module, K3S) {
			defaultVersion, err := kubernetesversions.Default(client, clusters.K3SClusterType.String(), nil)
			require.NoError(t, err)

			clusterConfig.KubernetesVersion = defaultVersion[0]
		}
	}
}

// DefaultUpgradedK8sVersion is a function that will set the default Kubernetes upgrade version
// if the user has not specified one.
func DefaultUpgradedK8sVersion(t *testing.T, client *rancher.Client, clusterConfig *config.TerratestConfig, terraformConfig *config.TerraformConfig) {
	if clusterConfig.UpgradedKubernetesVersion == "" {
		if strings.Contains(clusterConfig.KubernetesVersion, RANCHER) {
			defaultVersion, err := kubernetesversions.Default(client, clusters.RKE1ClusterType.String(), nil)
			require.NoError(t, err)

			clusterConfig.KubernetesVersion = defaultVersion[0]
		} else if strings.Contains(clusterConfig.KubernetesVersion, RKE2) {
			defaultVersion, err := kubernetesversions.Default(client, clusters.RKE2ClusterType.String(), nil)
			require.NoError(t, err)

			clusterConfig.KubernetesVersion = defaultVersion[0]
		} else if strings.Contains(clusterConfig.KubernetesVersion, K3S) {
			defaultVersion, err := kubernetesversions.Default(client, clusters.K3SClusterType.String(), nil)
			require.NoError(t, err)

			clusterConfig.KubernetesVersion = defaultVersion[0]
		}
	} else {
		clusterConfig.KubernetesVersion = clusterConfig.UpgradedKubernetesVersion
	}
}
