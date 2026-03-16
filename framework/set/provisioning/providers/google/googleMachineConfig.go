package google

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/google"
	"github.com/zclconf/go-cty/cty"
)

// SetGoogleRKE2K3SMachineConfig is a helper function that will set the Google RKE2/K3S
// Terraform machine configurations in the main.tf file.
func SetGoogleRKE2K3SMachineConfig(machineConfigBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	googleConfigBlock := machineConfigBlockBody.AppendNewBlock(google.GoogleConfig, nil)
	googleConfigBlockBody := googleConfigBlock.Body()

	googleConfigBlockBody.SetAttributeValue(google.DiskSize, cty.NumberIntVal(terraformConfig.GoogleConfig.DiskSize))
	googleConfigBlockBody.SetAttributeValue(google.DiskType, cty.StringVal(terraformConfig.GoogleConfig.DiskType))
	googleConfigBlockBody.SetAttributeValue(google.MachineImage, cty.StringVal(terraformConfig.GoogleConfig.MachineImage))
	googleConfigBlockBody.SetAttributeValue(google.MachineType, cty.StringVal(terraformConfig.GoogleConfig.MachineType))
	googleConfigBlockBody.SetAttributeValue(google.Network, cty.StringVal(terraformConfig.GoogleConfig.Network))
	googleConfigBlockBody.SetAttributeValue(google.Project, cty.StringVal(terraformConfig.GoogleConfig.ProjectID))
	googleConfigBlockBody.SetAttributeValue(google.Zone, cty.StringVal(terraformConfig.GoogleConfig.Zone))
}
