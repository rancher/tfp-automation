package azure

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/azure"
	"github.com/zclconf/go-cty/cty"
)

// CreateAzureNetworkInterface is a function that will set the network interface configurations in the main.tf file.
func CreateAzureNetworkInterface(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, hostnamePrefix string) {
	networkInterfaceBlock := rootBody.AppendNewBlock(general.Resource, []string{azure.AzureNetworkInterface, azure.AzureNetworkInterface + "-" + hostnamePrefix})
	networkInterfaceBlockBody := networkInterfaceBlock.Body()

	networkInterfaceBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+"-nic-"+hostnamePrefix))

	expression := azure.AzureResourceGroup + "." + azure.AzureResourceGroup + ".location"
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	networkInterfaceBlockBody.SetAttributeRaw(location, values)

	expression = azure.AzureResourceGroup + "." + azure.AzureResourceGroup + ".name"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	networkInterfaceBlockBody.SetAttributeRaw(resourceGroupName, values)

	ipConfigurationBlock := networkInterfaceBlockBody.AppendNewBlock(ipConfiguration, nil)
	ipConfigurationBlockBody := ipConfigurationBlock.Body()

	ipConfigurationBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal("internal"))

	expression = azure.AzureSubnet + "." + azure.AzureSubnet + ".id"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	ipConfigurationBlockBody.SetAttributeRaw(subnetID, values)
	ipConfigurationBlockBody.SetAttributeValue(privateIPAddressAllocation, cty.StringVal("Dynamic"))

	expression = azure.AzurePublicIP + "." + azure.AzurePublicIP + "-" + hostnamePrefix + ".id"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	ipConfigurationBlockBody.SetAttributeRaw(publicIPAddressID, values)
}
