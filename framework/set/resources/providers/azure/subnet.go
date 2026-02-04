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

// CreateAzureSubnet is a function that will set the subnet configurations in the main.tf file.
func CreateAzureSubnet(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	subnetBlock := rootBody.AppendNewBlock(general.Resource, []string{azure.AzureSubnet, azure.AzureSubnet})
	subnetBlockBody := subnetBlock.Body()

	subnetBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+"-subnet"))

	expression := azure.AzureResourceGroup + "." + azure.AzureResourceGroup + ".name"
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	subnetBlockBody.SetAttributeRaw(resourceGroupName, values)

	expression = azure.AzureVirtualNetwork + "." + azure.AzureVirtualNetwork + ".name"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	subnetBlockBody.SetAttributeRaw(virtualNetworkName, values)

	prefixes := format.ListOfStrings([]string{"10.0.1.0/24"})
	subnetBlockBody.SetAttributeRaw(addressPrefixes, prefixes)
}
