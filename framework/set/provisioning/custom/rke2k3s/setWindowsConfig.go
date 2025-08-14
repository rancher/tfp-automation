package rke2k3s

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/nullresource"
)

// SetCustomRKE2Windows is a function that will set the custom RKE2 cluster configurations in the main.tf file.
func SetCustomRKE2Windows(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, configMap []map[string]any,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) (*hclwrite.File, *os.File, error) {
	nullresource.CustomWindowsNullResource(rootBody, terraformConfig, terraformConfig.ResourcePrefix)
	rootBody.AppendNewline()

	return newFile, file, nil
}
