package provisioning

import (
	"strings"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/clusters/kubernetesversions"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
)

// GetK8sVersion is a function that will set the Kubernetes version if the user has not specified one. It
// will get the latest version or the second latest version based on the versionType.
func GetK8sVersion(client *rancher.Client, cattleConfig map[string]any) error {
	terraform := new(config.TerraformConfig)
	operations.LoadObjectFromMap(config.TerraformConfigurationFileKey, cattleConfig, terraform)

	terratest := new(config.TerratestConfig)
	operations.LoadObjectFromMap(config.TerratestConfigurationFileKey, cattleConfig, terratest)

	defaultVersion := terratest.KubernetesVersion
	if terratest.KubernetesVersion == "" {
		if strings.Contains(terraform.Module, clustertypes.RKE2) {
			defaultVersions, err := kubernetesversions.Default(client, clustertypes.RKE2, nil)
			if err != nil {
				return err
			}

			defaultVersion = defaultVersions[0]
		} else if strings.Contains(terraform.Module, clustertypes.K3S) {
			defaultVersions, err := kubernetesversions.Default(client, clustertypes.K3S, nil)
			if err != nil {
				return err
			}

			defaultVersion = defaultVersions[0]
		}
	}

	operations.ReplaceValue([]string{"terratest", "kubernetesVersion"}, defaultVersion, cattleConfig)
	return nil
}
