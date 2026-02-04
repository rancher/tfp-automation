package azure

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/format"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/azure"
	"github.com/zclconf/go-cty/cty"
)

// CreateAzureVirtualNetwork is a function that will set the virtual network configurations in the main.tf file.
func CreateAzureVirtualNetwork(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	virtualNetworkBlock := rootBody.AppendNewBlock(general.Resource, []string{azure.AzureVirtualNetwork, azure.AzureVirtualNetwork})
	virtualNetworkBlockBody := virtualNetworkBlock.Body()

	virtualNetworkBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+"-vnet"))

	expression := azure.AzureResourceGroup + "." + azure.AzureResourceGroup + ".name"
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	virtualNetworkBlockBody.SetAttributeRaw(resourceGroupName, values)

	expression = azure.AzureResourceGroup + "." + azure.AzureResourceGroup + ".location"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	virtualNetworkBlockBody.SetAttributeRaw(location, values)

	prefixes := format.ListOfStrings([]string{"10.0.0.0/16"})
	virtualNetworkBlockBody.SetAttributeRaw(addressSpace, prefixes)
}
