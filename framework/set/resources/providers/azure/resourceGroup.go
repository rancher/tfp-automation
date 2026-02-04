package azure

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/azure"
	"github.com/zclconf/go-cty/cty"
)

// CreateAzureResourceGroup is a function that will set the resource group configurations in the main.tf file.
func CreateAzureResourceGroup(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	resourceBlock := rootBody.AppendNewBlock(general.Resource, []string{azure.AzureResourceGroup, azure.AzureResourceGroup})
	resourceBlockBody := resourceBlock.Body()

	resourceBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+"-rg"))
	resourceBlockBody.SetAttributeValue(location, cty.StringVal(terraformConfig.AzureConfig.Location))
}
