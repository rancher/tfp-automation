package set

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/framework/set/provisioning/hosted"
)

// HostedClusters is a function that will set the hosted clusters in the main.tf file.
func HostedClusters(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, newFile *hclwrite.File,
	rootBody *hclwrite.Body, file *os.File) (*hclwrite.File, *os.File, error) {
	var err error

	if strings.Contains(terraformConfig.Module, clustertypes.AKS) {
		newFile, file, err = hosted.SetAKS(terraformConfig, terratestConfig, newFile, rootBody, file)
		if err != nil {
			return newFile, file, err
		}
	}

	if strings.Contains(terraformConfig.Module, clustertypes.EKS) {
		newFile, file, err = hosted.SetEKS(terraformConfig, terratestConfig, newFile, rootBody, file)
		if err != nil {
			return newFile, file, err
		}
	}

	if strings.Contains(terraformConfig.Module, clustertypes.GKE) {
		newFile, file, err = hosted.SetGKE(terraformConfig, terratestConfig, newFile, rootBody, file)
		if err != nil {
			return newFile, file, err
		}
	}

	return newFile, file, nil
}
