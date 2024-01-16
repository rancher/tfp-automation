package linode

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/zclconf/go-cty/cty"
)

// SetLinodeRKE2K3SMachineConfig is a helper function that will set the EC2 RKE2/K3S Terraform machine configurations in the main.tf file.
func SetLinodeRKE2K3SMachineConfig(machineConfigBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	linodeConfigBlock := machineConfigBlockBody.AppendNewBlock("linode_config", nil)
	linodeConfigBlockBody := linodeConfigBlock.Body()

	linodeConfigBlockBody.SetAttributeValue("image", cty.StringVal(terraformConfig.LinodeConfig.LinodeImage))
	linodeConfigBlockBody.SetAttributeValue("region", cty.StringVal(terraformConfig.LinodeConfig.Region))
	linodeConfigBlockBody.SetAttributeValue("root_pass", cty.StringVal(terraformConfig.LinodeConfig.LinodeRootPass))
	linodeConfigBlockBody.SetAttributeValue("token", cty.StringVal(terraformConfig.LinodeConfig.LinodeToken))
}
