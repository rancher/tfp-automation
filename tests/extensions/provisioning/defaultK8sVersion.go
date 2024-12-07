package provisioning

import (
	"strings"
	"testing"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/clusters/kubernetesversions"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// GetK8sVersion is a function that will set the Kubernetes version if the user has not specified one. It
// will get the latest version or the second latest version based on the versionType.
func GetK8sVersion(t *testing.T, client *rancher.Client, terratestConfig *config.TerratestConfig, terraformConfig *config.TerraformConfig, versionType string) {
	var defaultVersion []string
	var err error

	if terratestConfig.KubernetesVersion == "" {
		switch versionType {
		case configs.DefaultK8sVersion:
			if strings.Contains(terraformConfig.Module, clustertypes.RKE1) {
				defaultVersion, err = kubernetesversions.Default(client, clusters.RKE1ClusterType.String(), nil)
				require.NoError(t, err)

				terratestConfig.KubernetesVersion = defaultVersion[0]
			} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) {
				defaultVersion, err = kubernetesversions.Default(client, clusters.RKE2ClusterType.String(), nil)
				require.NoError(t, err)

				terratestConfig.KubernetesVersion = defaultVersion[0]
			} else if strings.Contains(terraformConfig.Module, clustertypes.K3S) {
				defaultVersion, err = kubernetesversions.Default(client, clusters.K3SClusterType.String(), nil)
				require.NoError(t, err)

				terratestConfig.KubernetesVersion = defaultVersion[0]
			}
		case configs.SecondHighestVersion:
			if strings.Contains(terraformConfig.Module, clustertypes.RKE1) {
				defaultVersion, err = kubernetesversions.ListRKE1AllVersions(client)
				require.NoError(t, err)

				terratestConfig.KubernetesVersion = defaultVersion[2]
			} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) {
				defaultVersion, err = kubernetesversions.ListRKE2AllVersions(client)
				require.NoError(t, err)

				terratestConfig.KubernetesVersion = defaultVersion[2]
			} else if strings.Contains(terraformConfig.Module, clustertypes.K3S) {
				defaultVersion, err = kubernetesversions.ListK3SAllVersions(client)
				require.NoError(t, err)

				terratestConfig.KubernetesVersion = defaultVersion[2]
			}
		default:
			logrus.Errorf("Invalid version type: %s", versionType)
		}
	}
}

// DefaultUpgradedK8sVersion is a function that will set the default Kubernetes upgrade version
// if the user has not specified one.
func DefaultUpgradedK8sVersion(t *testing.T, client *rancher.Client, terratestConfig *config.TerratestConfig, terraformConfig *config.TerraformConfig) {
	if terratestConfig.UpgradedKubernetesVersion == "" {
		if strings.Contains(terratestConfig.KubernetesVersion, clustertypes.RANCHER) {
			defaultVersion, err := kubernetesversions.Default(client, clusters.RKE1ClusterType.String(), nil)
			require.NoError(t, err)

			terratestConfig.KubernetesVersion = defaultVersion[0]
		} else if strings.Contains(terratestConfig.KubernetesVersion, clustertypes.RKE2) {
			defaultVersion, err := kubernetesversions.Default(client, clusters.RKE2ClusterType.String(), nil)
			require.NoError(t, err)

			terratestConfig.KubernetesVersion = defaultVersion[0]
		} else if strings.Contains(terratestConfig.KubernetesVersion, clustertypes.K3S) {
			defaultVersion, err := kubernetesversions.Default(client, clusters.K3SClusterType.String(), nil)
			require.NoError(t, err)

			terratestConfig.KubernetesVersion = defaultVersion[0]
		}
	} else {
		terratestConfig.KubernetesVersion = terratestConfig.UpgradedKubernetesVersion
	}
}
