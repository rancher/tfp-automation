package provisioning

import (
	"slices"
	"strings"
	"testing"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/clusters/kubernetesversions"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// GetK8sVersion is a function that will set the Kubernetes version if the user has not specified one. It
// will get the latest version or the second latest version based on the versionType.
func GetK8sVersion(t *testing.T, client *rancher.Client, terratestConfig *config.TerratestConfig, terraformConfig *config.TerraformConfig,
	versionType string, configMap []map[string]any) {
	var defaultVersion string

	terraform := new(config.TerraformConfig)
	operations.LoadObjectFromMap(config.TerraformConfigurationFileKey, configMap[0], terraform)

	terratest := new(config.TerratestConfig)
	operations.LoadObjectFromMap(config.TerratestConfigurationFileKey, configMap[0], terratest)

	if terratest.KubernetesVersion == "" {
		switch versionType {
		case configs.DefaultK8sVersion:
			if strings.Contains(terraform.Module, clustertypes.RKE1) {
				defaultVersions, err := kubernetesversions.Default(client, clusters.RKE1ClusterType.String(), nil)
				require.NoError(t, err)

				defaultVersion = defaultVersions[0]
			} else if strings.Contains(terraform.Module, clustertypes.RKE2) {
				defaultVersions, err := kubernetesversions.Default(client, clusters.RKE2ClusterType.String(), nil)
				require.NoError(t, err)

				defaultVersion = defaultVersions[0]
			} else if strings.Contains(terraform.Module, clustertypes.K3S) {
				defaultVersions, err := kubernetesversions.Default(client, clusters.K3SClusterType.String(), nil)
				require.NoError(t, err)

				defaultVersion = defaultVersions[0]
			}
		case configs.SecondHighestVersion:
			if strings.Contains(terraform.Module, clustertypes.RKE1) {
				defaultVersions, err := kubernetesversions.ListRKE1AllVersions(client)
				require.NoError(t, err)

				slices.Reverse(defaultVersions)

				defaultVersion = defaultVersions[1]
			} else if strings.Contains(terraform.Module, clustertypes.RKE2) {
				defaultVersions, err := kubernetesversions.ListRKE2AllVersions(client)
				require.NoError(t, err)

				slices.Reverse(defaultVersions)

				defaultVersion = defaultVersions[1]
			} else if strings.Contains(terraform.Module, clustertypes.K3S) {
				defaultVersions, err := kubernetesversions.ListK3SAllVersions(client)
				require.NoError(t, err)

				slices.Reverse(defaultVersions)

				defaultVersion = defaultVersions[1]
			}
		default:
			logrus.Errorf("Invalid version type: %s", versionType)
		}
	} else {
		defaultVersion = terratest.KubernetesVersion
	}

	operations.ReplaceValue([]string{"terratest", "kubernetesVersion"}, defaultVersion, configMap[0])
}

// DefaultUpgradedK8sVersion is a function that will set the default Kubernetes upgrade version
// if the user has not specified one.
func DefaultUpgradedK8sVersion(t *testing.T, client *rancher.Client, terratestConfig *config.TerratestConfig, terraformConfig *config.TerraformConfig,
	configMap []map[string]any) {
	var defaultVersion string

	if terratestConfig.UpgradedKubernetesVersion == "" {
		switch {
		case strings.Contains(terratestConfig.KubernetesVersion, clustertypes.RANCHER):
			defaultVersions, err := kubernetesversions.Default(client, clusters.RKE1ClusterType.String(), nil)
			require.NoError(t, err)

			defaultVersion = defaultVersions[0]
		case strings.Contains(terratestConfig.KubernetesVersion, clustertypes.RKE2):
			defaultVersions, err := kubernetesversions.Default(client, clusters.RKE2ClusterType.String(), nil)
			require.NoError(t, err)

			defaultVersion = defaultVersions[0]
		case strings.Contains(terratestConfig.KubernetesVersion, clustertypes.K3S):
			defaultVersions, err := kubernetesversions.Default(client, clusters.K3SClusterType.String(), nil)
			require.NoError(t, err)

			defaultVersion = defaultVersions[0]
		default:
			terratest := new(config.TerratestConfig)
			operations.LoadObjectFromMap(config.TerratestConfigurationFileKey, configMap[0], terratest)

			defaultVersion = terratest.UpgradedKubernetesVersion
		}
	} else {
		terratest := new(config.TerratestConfig)
		operations.LoadObjectFromMap(config.TerratestConfigurationFileKey, configMap[0], terratest)

		defaultVersion = terratest.UpgradedKubernetesVersion
	}

	operations.ReplaceValue([]string{"terratest", "kubernetesVersion"}, defaultVersion, configMap[0])
}
