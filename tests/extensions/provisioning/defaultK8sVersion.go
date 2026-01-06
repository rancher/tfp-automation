package provisioning

import (
	"strings"
	"testing"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/clusters/kubernetesversions"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/stretchr/testify/require"
)

// GetK8sVersion is a function that will set the Kubernetes version if the user has not specified one. It
// will get the latest version or the second latest version based on the versionType.
func GetK8sVersion(t *testing.T, client *rancher.Client, terratestConfig *config.TerratestConfig, terraformConfig *config.TerraformConfig,
	configMap []map[string]any) {
	var defaultVersion string

	terraform := new(config.TerraformConfig)
	operations.LoadObjectFromMap(config.TerraformConfigurationFileKey, configMap[0], terraform)

	terratest := new(config.TerratestConfig)
	operations.LoadObjectFromMap(config.TerratestConfigurationFileKey, configMap[0], terratest)

	if terratest.KubernetesVersion == "" {
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
	} else {
		defaultVersion = terratest.KubernetesVersion
	}

	operations.ReplaceValue([]string{"terratest", "kubernetesVersion"}, defaultVersion, configMap[0])
}
