package linode

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/linode"
	"github.com/zclconf/go-cty/cty"
)

const (
	cloudCredential = "rancher2_cloud_credential"

	region       = "region"
	resource     = "resource"
	resourceName = "name"
)

// SetLinodeRKE2K3SMachineConfig is a helper function that will set the Linode RKE2/K3S
// Terraform machine configurations in the main.tf file.
func SetLinodeRKE2K3SMachineConfig(machineConfigBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	linodeConfigBlock := machineConfigBlockBody.AppendNewBlock(linode.LinodeConfig, nil)
	linodeConfigBlockBody := linodeConfigBlock.Body()

	linodeConfigBlockBody.SetAttributeValue(linode.Image, cty.StringVal(terraformConfig.LinodeConfig.LinodeImage))
	linodeConfigBlockBody.SetAttributeValue(region, cty.StringVal(terraformConfig.LinodeConfig.Region))
	linodeConfigBlockBody.SetAttributeValue(linode.RootPass, cty.StringVal(terraformConfig.LinodeConfig.LinodeRootPass))
}
