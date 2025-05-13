package vsphere

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// CreateVsphereVirtualMachineTemplate is a function that will set the vSphere virtual machine template configuration in the main.tf file.
func CreateVsphereVirtualMachineTemplate(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, dataCenterValue hclwrite.Tokens) {
	vmTemplateBlock := rootBody.AppendNewBlock(defaults.Data, []string{defaults.VsphereVirtualMachine, defaults.VsphereVirtualMachineTemplate})
	vmTemplateBlockBody := vmTemplateBlock.Body()

	vmTemplateBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.VsphereConfig.CloneFrom))
	vmTemplateBlockBody.SetAttributeRaw(datacenterID, dataCenterValue)
}
