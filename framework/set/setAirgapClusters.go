package set

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/provisioning/airgap"
)

// AirgapClusters is a function that will set the airgap clusters in the main.tf file.
func AirgapClusters(client *rancher.Client, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File, configMap []map[string]any,
	isWindows bool) (*hclwrite.File, *os.File, error) {
	var err error

	if !isWindows {
		newFile, file, err = airgap.SetAirgapRKE2K3s(terraformConfig, terratestConfig, configMap, newFile, rootBody, file)
		if err != nil {
			return newFile, file, err
		}
	}

	if isWindows {
		newFile, file, err = airgap.SetAirgapRKE2Windows(terraformConfig, terratestConfig, configMap, newFile, rootBody, file)
		if err != nil {
			return newFile, file, err
		}
	}

	return newFile, file, nil
}
