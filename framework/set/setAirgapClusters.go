package set

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/framework/set/provisioning/airgap/rke1"
	"github.com/rancher/tfp-automation/framework/set/provisioning/airgap/rke2k3s"
)

// AirgapClusters is a function that will set the airgap clusters in the main.tf file.
func AirgapClusters(client *rancher.Client, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File, configMap []map[string]any,
	isWindows bool) (*hclwrite.File, *os.File, error) {
	var err error

	if strings.Contains(terraformConfig.Module, clustertypes.RKE1) {
		newFile, file, err = rke1.SetAirgapRKE1(terraformConfig, terratestConfig, configMap, newFile, rootBody, file)
		if err != nil {
			return newFile, file, err
		}
	}

	if strings.Contains(terraformConfig.Module, clustertypes.RKE2) || strings.Contains(terraformConfig.Module, clustertypes.K3S) {
		if !isWindows {
			newFile, file, err = rke2k3s.SetAirgapRKE2K3s(terraformConfig, terratestConfig, configMap, newFile, rootBody, file)
			if err != nil {
				return newFile, file, err
			}
		}

		if isWindows {
			newFile, file, err = rke2k3s.SetAirgapRKE2Windows(terraformConfig, terratestConfig, configMap, newFile, rootBody, file)
			if err != nil {
				return newFile, file, err
			}
		}
	}

	return newFile, file, nil
}
